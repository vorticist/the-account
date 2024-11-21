package handlers

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/lithammer/shortuuid/v4"
	"github.com/vorticist/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"html/template"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"vortex.studio/account/internal/repo"
	"vortex.studio/account/internal/structs"
	"vortex.studio/account/internal/utils"
)

var (
	store = sessions.NewCookieStore([]byte("super-secret-key"))
)

var templateFuncs = template.FuncMap{
	"makeURLSafe": makeURLSafe,
}

type Handler struct {
	venueRepo        *repo.VenueRepository
	activeTablesRepo *repo.ActiveTablesRepository
}

func NewHandler(repository repo.VenueRepository, activeTablesRepo *repo.ActiveTablesRepository) *Handler {
	return &Handler{
		venueRepo:        &repository,
		activeTablesRepo: activeTablesRepo,
	}
}

func (h *Handler) AccountHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")

	// Check if user is authenticated
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		tmpl := template.Must(template.ParseFiles("templates/login.html"))
		tmpl.Execute(w, nil)
		return
	}

	venues, err := h.venueRepo.GetAllVenues(r.Context())
	if err != nil {
		logger.Errorf("error fetching venues: %v", err)
		http.Error(w, "Error fetching venues", http.StatusInternalServerError)
		return
	}
	adminPage := structs.AdminPage{
		Title:  "Table Codes",
		Venues: venues,
	}
	tmpl := template.Must(template.New("admin.html").Funcs(templateFuncs).ParseFiles("templates/admin.html"))
	err = tmpl.Execute(w, adminPage)
	if err != nil {
		logger.Errorf("error executing template: %v", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) AddTableHandler(w http.ResponseWriter, r *http.Request) {
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
		Name:        "Venue 1",
		Description: "Description 1",
		Image:       "image1.jpg",
		TableCodes:  []structs.TableCode{},
	}
	generateTableCodes(&venue, tables)

	adminPage := structs.AdminPage{
		Title:  "Table Codes",
		Venues: []structs.Venue{venue},
	}

	w.WriteHeader(http.StatusOK)
	tmpl := template.New("venue-list.html").Funcs(templateFuncs)
	tmpl = template.Must(tmpl.ParseFiles("templates/venue-list.html"))
	tmpl.Execute(w, adminPage)
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = false
	session.Save(r, w)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) VenueHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the form data into a Venue struct
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Venue name is required", http.StatusBadRequest)
		return
	}

	numberOfTablesStr := r.FormValue("numberOfTables")
	numberOfTables, err := strconv.Atoi(numberOfTablesStr)
	if err != nil {
		http.Error(w, "Number of tables must be a valid number", http.StatusBadRequest)
		return
	}
	venue := structs.Venue{
		Name: name,
	}

	if numberOfTables > 0 {
		generateTableCodes(&venue, numberOfTables)

	}

	// Save the venue to the database
	if _, err := h.venueRepo.CreateVenue(&venue); err != nil {
		logger.Errorf("error creating venue: %v", err)
		http.Error(w, "Failed to create venue", http.StatusInternalServerError)
		return
	}
	venues, err := h.venueRepo.GetAllVenues(r.Context())
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		logger.Errorf("error fetching venues: %v", err)
		http.Error(w, "Error fetching venues", http.StatusInternalServerError)
		return
	}
	adminPage := structs.AdminPage{
		Title:  "Table Codes",
		Venues: venues,
	}

	w.WriteHeader(http.StatusOK)
	tmpl := template.Must(template.New("venue-list.html").Funcs(templateFuncs).ParseFiles("templates/venue-list.html"))
	tmpl.Execute(w, adminPage)
}

func (h *Handler) CodeHandler(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]
	logger.Infof("got code: %v", code)

	// Check for existing client cookie
	clientID := ""
	cookie, err := r.Cookie("client_id")
	if err == http.ErrNoCookie {
		// Generate new client ID if cookie doesn't exist
		clientID = uuid.New().String()
		http.SetCookie(w, &http.Cookie{
			Name:     "client_id",
			Value:    clientID,
			Path:     "/",
			MaxAge:   86400 * 30, // 30 days
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
	} else {
		clientID = cookie.Value
	}

	logger.Infof("client ID: %v", clientID)
	session, err := h.activeTablesRepo.GetSessionForTable(code)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		logger.Errorf("error fetching session: %v", err)
		http.Error(w, "Error fetching session", http.StatusInternalServerError)
		return
	}

	if session == nil {
		session = &structs.ActiveTables{
			ClientID:  clientID,
			TableCode: code,
		}
		_, err = h.activeTablesRepo.TableActive(session)
		if err != nil {
			logger.Errorf("error creating session: %v", err)
			http.Error(w, "Error creating session", http.StatusInternalServerError)
			return
		}
	}

	if session.ClientID != clientID {
		tmpl := template.Must(template.New("occupied.html").Funcs(templateFuncs).ParseFiles("templates/occupied.html"))
		tmpl.Execute(w, nil)
		return
	}

	venue, err := h.venueRepo.GetVenueByTableCode(r.Context(), code)
	if err != nil {
		logger.Errorf("error fetching venue: %v", err)
		http.Error(w, "Error fetching venue", http.StatusInternalServerError)
		return
	}
	if venue == nil {
		logger.Errorf("venue not found for code: %v - %v", code, err)
		http.Error(w, "Venue not found", http.StatusNotFound)
		return
	}

	logger.Infof("found venue: %v", venue)
	menu, err := h.venueRepo.GetMenuForVenue(r.Context(), venue.ID)
	if err != nil {
		logger.Errorf("error fetching menu: %v", err)
		http.Error(w, "Error fetching menu", http.StatusInternalServerError)
		return
	}
	logger.Infof("found menu: %v", menu)

	menuPage := structs.MenuPage{
		Title: "Menu",
		Menu:  *menu,
	}
	tmpl := template.Must(template.New("menu.html").Funcs(templateFuncs).ParseFiles("templates/menu.html"))
	err = tmpl.Execute(w, menuPage)
	if err != nil {
		logger.Errorf("error executing template: %v", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
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

func generateTableCodes(venue *structs.Venue, howMany int) error {
	venue.TableCodes = []structs.TableCode{}
	for i := 0; i < howMany; i++ {
		u := shortuuid.New()
		tableCodeUrl := fmt.Sprintf("http://localhost:9090/table/%s", u)
		qrCode, err := utils.GenerateQRCodeBase64(tableCodeUrl)
		if err != nil {
			return err
		}
		venue.TableCodes = append(venue.TableCodes, structs.TableCode{
			Code:    u,
			CodeUrl: tableCodeUrl,
			Base64:  qrCode,
		})
	}
	return nil
}
