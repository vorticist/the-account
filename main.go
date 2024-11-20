package main

import (
	"encoding/base64"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/lithammer/shortuuid/v4"
	"github.com/vorticist/logger"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"vortex.studio/account/internal/structs"

	"github.com/skip2/go-qrcode"
)

var (
	store = sessions.NewCookieStore([]byte("super-secret-key"))
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/admin", accountHandler).Methods("GET")
	router.HandleFunc("/add-table", addTableHandler).Methods("POST")
	router.HandleFunc("/login", loginHandler).Methods("POST")
	router.HandleFunc("/logout", logoutHandler).Methods("GET")

	router.HandleFunc("/{code}", getCodeHandler).Methods("GET", "POST")

	log.Fatal(http.ListenAndServe(":9090", router))
}

func addTableHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	tableCount := r.FormValue("tableCount")
	if _, err := strconv.Atoi(tableCount); err != nil {
		http.Error(w, "Table count must be a number", http.StatusBadRequest)
		return
	}
	tables, _ := strconv.Atoi(tableCount)

	venue := structs.Venue{
		ID:          1,
		Name:        "Venue 1",
		Description: "Description 1",
		Image:       "image1.jpg",
		TableCodes:  []structs.TableCode{},
	}

	for _ = range tables {
		u := shortuuid.New()
		tableCodeUrl := fmt.Sprintf("http://localhost:9090/%s", u)
		qrCode, err := generateQRCodeBase64(tableCodeUrl)
		if err != nil {
			http.Error(w, "Error generating QR code", http.StatusInternalServerError)
			logger.Errorf("error generating QR code: %s", err)
			return
		}
		venue.TableCodes = append(venue.TableCodes, structs.TableCode{
			Code:   tableCodeUrl,
			Base64: qrCode,
		})
	}

	adminPage := structs.AdminPage{
		Title: "Table Codes",
		Venue: venue,
	}

	w.WriteHeader(http.StatusOK)
	tmpl := template.Must(template.ParseFiles("templates/venue-list.html"))
	tmpl.Execute(w, adminPage)
}

func makeIndexHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info(" admin handler")
		tmpl := template.Must(template.ParseFiles("templates/base.html"))
		tmpl.Execute(w, nil)
	}
}

func getCodeHandler(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]
	fmt.Println(code)
}

func accountHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	// Check if user is authenticated
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		tmpl := template.Must(template.ParseFiles("templates/login.html"))
		tmpl.Execute(w, nil)
		return
	}

	// User is authenticated, show index page
	tmpl := template.Must(template.ParseFiles("templates/admin.html"))
	venue := structs.Venue{
		ID:          1,
		Name:        "Venue 1",
		Description: "Description 1",
		Image:       "image1.jpg",
		TableCodes:  []structs.TableCode{},
	}
	adminPage := structs.AdminPage{
		Title: "Table Codes",
		Venue: venue,
	}
	tmpl.Execute(w, adminPage)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	// Here you would validate credentials
	// This is a simplified example
	username := r.FormValue("username")
	password := r.FormValue("password")

	logger.Infof("loginHandler username: %s, password: %s", username, password)

	if username == "admin" && password == "password" {
		session.Values["authenticated"] = true
		session.Save(r, w)
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = false
	session.Save(r, w)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func generateQRCodeBase64(content string) (string, error) {
	// Generate the QR code as PNG data
	pngData, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}

	// Encode the PNG data as a Base64 string
	base64Data := base64.StdEncoding.EncodeToString(pngData)

	// Return the data with a proper data URI scheme for embedding in HTML or other contexts
	return base64Data, nil
}
func makeURLSafe(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")

	// Remove special characters and keep only alphanumeric and hyphens
	reg := regexp.MustCompile("[^a-z0-9-]+")
	name = reg.ReplaceAllString(name, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	name = reg.ReplaceAllString(name, "-")

	// Trim hyphens from start and end
	name = strings.Trim(name, "-")

	return name
}
