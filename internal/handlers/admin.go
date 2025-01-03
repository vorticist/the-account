package handlers

import (
	"errors"
	"github.com/vorticist/logger"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"html/template"
	"net/http"
	"strconv"
	"vortex.studio/account/internal/repo"
	"vortex.studio/account/internal/structs"

	menu "vortex.studio/account/internal/menu-analyzer"
)

type AdminHandler struct {
	venueRepo  *repo.VenueRepository
	tablesRepo *repo.ActiveTablesRepository
	menuRepo   *repo.MenuRepository
}

func NewAdminHandler(repository repo.VenueRepository, tablesRepo *repo.ActiveTablesRepository, menuRepo *repo.MenuRepository) *AdminHandler {
	return &AdminHandler{
		venueRepo:  &repository,
		tablesRepo: tablesRepo,
		menuRepo:   menuRepo,
	}
}

func (h *AdminHandler) AccountHandler(w http.ResponseWriter, r *http.Request) {
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

	openSessions, err := h.tablesRepo.GetOpenSessions(r.Context())
	if err != nil {
		logger.Errorf("error fetching open sessions: %v", err)
		http.Error(w, "Error fetching open sessions", http.StatusInternalServerError)
		return
	}

	adminPage := structs.AdminPage{
		Title:        "Table Codes",
		Venues:       venues,
		OpenSessions: openSessions,
	}
	tmpl := template.Must(template.New("admin.html").Funcs(templateFuncs).ParseFiles("templates/admin.html"))
	err = tmpl.Execute(w, adminPage)
	if err != nil {
		logger.Errorf("error executing template: %v", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

func (h *AdminHandler) AddTableHandler(w http.ResponseWriter, r *http.Request) {
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

func (h *AdminHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
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

func (h *AdminHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	session.Values["authenticated"] = false
	session.Save(r, w)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *AdminHandler) VenueHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	auth, ok := session.Values["authenticated"].(bool)
	if !ok || !auth {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the form data into a Venue struct
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
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

	// Get file from form
	file, _, err := r.FormFile("menuFile")
	if err != nil {
		logger.Errorf("error getting file from form: %v", err)
		http.Error(w, "Error getting file from form", http.StatusBadRequest)
		return
	}
	ar := menu.StartMenuFileAnalysis(file)
	menuAnalysisResult := <-ar
	if menuAnalysisResult.Err != nil {
		logger.Errorf("error analyzing menu file: %v", menuAnalysisResult.Err)
		http.Error(w, "Error analyzing menu file", http.StatusInternalServerError)
		return
	}

	venue := structs.Venue{
		Name: name,
	}

	if numberOfTables > 0 {
		generateTableCodes(&venue, numberOfTables)

	}

	// Save the venue to the database
	insertResult, err := h.venueRepo.CreateVenue(&venue)
	if err != nil {
		logger.Errorf("error creating venue: %v", err)
		http.Error(w, "Failed to create venue", http.StatusInternalServerError)
		return
	}
	venueID, ok := insertResult.InsertedID.(primitive.ObjectID)
	if !ok {
		logger.Errorf("error converting InsertedID to ObjectID")
		http.Error(w, "Error creating venue", http.StatusInternalServerError)
		return
	}
	menuAnalysisResult.Result.VenueId = venueID
	_, err = h.menuRepo.CreateMenu(menuAnalysisResult.Result)
	if err != nil {
		logger.Errorf("error creating menu: %v", err)
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
