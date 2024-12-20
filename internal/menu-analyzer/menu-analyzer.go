package menu_analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/vorticist/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
	"strings"
	"vortex.studio/account/internal/structs"

	openai "github.com/sashabaranov/go-openai"
)

func StartMenuFileAnalysis(file multipart.File) <-chan AnalysisResponse {
	logger.Info("StartMenuFileAnalysis starting")
	analysisChannel := make(chan AnalysisResponse)
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

type AnalysisResponse struct {
	Err    error
	Result *AnalysisData
}

type analysisMessage struct {
	File         multipart.File `json:"file"`
	AnalysisData AnalysisData

	aCh chan AnalysisResponse
	err error
}

type AnalysisData struct {
	ID                primitive.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
	VenueId           primitive.ObjectID     `json:"venueId,omitempty" bson:"venueId"`
	VisionResult      map[string]interface{} `json:"visionResult,omitempty" bson:"visionResult,omitempty"`
	RawCategoryResult string                 `json:"rawCategoryResult,omitempty" bson:"rawCategoryResult,omitempty"`
	CategoryResult    structs.MenuData       `json:"categoryResult,omitempty" bson:"categoryResult,omitempty"`
}

type stage func(am *analysisMessage) stage

func onError(am *analysisMessage) stage {
	if am.err != nil {
		logger.Errorf("stopping analysis due to error: %v", am.err.Error())
		logger.Errorf("analysis data: %v", am.AnalysisData)
		am.aCh <- AnalysisResponse{Err: am.err}
		return nil
	}

	return nil
}

func onSuccess(am *analysisMessage) stage {
	logger.Infof("analysis result: %v", am.AnalysisData)
	am.aCh <- AnalysisResponse{Result: &am.AnalysisData}
	return nil
}

func visionAnalysis(am *analysisMessage) stage {
	logger.Info("vision analysis")
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

	am.AnalysisData.VisionResult = result

	return categoryAnalysis
}

func categoryAnalysis(am *analysisMessage) stage {
	logger.Info("categoryAnalysis")
	jsonString, err := mapToJSON(am.AnalysisData.VisionResult)
	if err != nil {
		am.err = err
		return onError
	}
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	typeDef := getTypeDefinition(structs.MenuData{})
	logger.Infof("typeDef: %v", typeDef)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT4o,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello! Can you come up with categories for the items in this json list. Rework the original json struct to reflect these categories and as best as you can make it match the provided go struct, please omit any additional comments or explanations and return the raw json struct.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: typeDef,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: jsonString,
				},
			},
		},
	)

	if err != nil {
		logger.Printf("ChatCompletion error: %v\n", err)
		am.err = err
		return onError
	}

	am.AnalysisData.RawCategoryResult = resp.Choices[0].Message.Content
	return mapCategoryResult
}

func mapCategoryResult(am *analysisMessage) stage {
	logger.Info("mapCategoryResult")
	var categoryMap structs.MenuData
	am.AnalysisData.RawCategoryResult = strings.Replace(am.AnalysisData.RawCategoryResult, "```json", "", -1)
	am.AnalysisData.RawCategoryResult = strings.Replace(am.AnalysisData.RawCategoryResult, "```", "", -1)

	err := json.Unmarshal([]byte(am.AnalysisData.RawCategoryResult), &categoryMap)
	if err != nil {
		am.err = err
		return onError
	}

	am.AnalysisData.CategoryResult = categoryMap

	return onSuccess
}

// mapToJSON converts a map[string]interface{} to its JSON string representation.
// It returns the JSON string and any error encountered during the process.
func mapToJSON(input map[string]interface{}) (string, error) {
	// Marshal the map into a JSON byte slice
	jsonBytes, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "", err
	}

	// Convert the byte slice to a string and return
	return string(jsonBytes), nil
}

func getTypeDefinition(i interface{}) string {
	visited := make(map[reflect.Type]bool)
	return getTypeDefinitionRecursive(i, visited)
}

func getTypeDefinitionRecursive(i interface{}, visited map[reflect.Type]bool) string {
	// Get the type of the passed argument
	t := reflect.TypeOf(i)

	// If we've already processed this type, avoid recursion (for cyclic structs)
	if visited[t] {
		return t.Name()
	}

	visited[t] = true

	// If it's a struct, generate a string representation of its fields
	if t.Kind() == reflect.Struct {
		var sb strings.Builder
		sb.WriteString("type " + t.Name() + " struct {\n")

		// Iterate through the fields of the struct
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			// Write the field name and its type
			sb.WriteString(fmt.Sprintf("\t%s %s", field.Name, getFieldTypeDefinition(field.Type, visited)))

			// Add a newline after each field
			sb.WriteString("\n")
		}

		sb.WriteString("}")
		return sb.String()
	}

	// If it's a pointer, dereference it
	if t.Kind() == reflect.Ptr {
		return getTypeDefinitionRecursive(reflect.New(t.Elem()).Interface(), visited)
	}

	// Handle slices and arrays
	if t.Kind() == reflect.Slice {
		return "[]" + getTypeDefinitionRecursive(reflect.New(t.Elem()).Interface(), visited)
	}

	// Handle maps
	if t.Kind() == reflect.Map {
		return "map[" + getTypeDefinitionRecursive(reflect.New(t.Key()).Interface(), visited) + "]" + getTypeDefinitionRecursive(reflect.New(t.Elem()).Interface(), visited)
	}

	// Return the type name for basic types
	return t.String()
}

func getFieldTypeDefinition(t reflect.Type, visited map[reflect.Type]bool) string {
	// If it's a nested struct, recurse into its definition
	if t.Kind() == reflect.Struct {
		return getTypeDefinitionRecursive(reflect.New(t).Interface(), visited)
	}

	return t.String()
}
