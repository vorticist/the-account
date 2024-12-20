package menu_analyzer

import (
	"bytes"
	"encoding/json"
	"github.com/vorticist/logger"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func StartMenuFileAnalysis(file multipart.File) <-chan string {
	analysisChannel := make(chan string)
	am := analysisMessage{
		File: file,
		aCh:  analysisChannel,
	}
	go func() {
		defer close(am.aCh)
		for stage := visionAnalysis; stage != nil; {
			stage = stage(&am)
		}
	}()
	return analysisChannel
}

type analysisMessage struct {
	File         multipart.File `json:"file"`
	VisionResult map[string]interface{}

	aCh chan string
	err error
}

type stage func(am *analysisMessage) stage

func onError(am *analysisMessage) stage {
	if am.err != nil {
		logger.Errorf("stopping analysis due to error: %v", am.err)
		am.aCh <- am.err.Error()
		return nil
	}

	return nil
}

func onSuccess(am *analysisMessage) stage {
	logger.Infof("analysis result: %v", am.VisionResult)
	am.aCh <- "Analysis completed successfully"
	return nil
}

func visionAnalysis(am *analysisMessage) stage {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "menu.jpg")
	if err != nil {
		am.err = err
		return onError
	}

	_, err = io.Copy(part, am.File)
	if err != nil {
		am.err = err
		return onError
	}

	err = writer.Close()
	if err != nil {
		am.err = err
		return onError
	}

	req, err := http.NewRequest("POST", os.Getenv("MENU_ANALYZER_URL"), body)
	if err != nil {
		am.err = err
		return onError
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		am.err = err
		return onError
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		am.err = err
		return onError
	}

	am.VisionResult = result

	return categoryAnalysis
}

func categoryAnalysis(am *analysisMessage) stage {

	return onSuccess
}
