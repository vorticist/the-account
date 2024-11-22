package handlers

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/vorticist/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"html/template"
	"net/http"
	"strconv"
	"vortex.studio/account/internal/repo"
	"vortex.studio/account/internal/structs"
)

type TableHandler struct {
	tablesRepo *repo.ActiveTablesRepository
	venuesRepo *repo.VenueRepository
	eventsRepo *repo.EventsRepo
}

func NewTablesHandler(venueRepo *repo.VenueRepository, activeTablesRepo *repo.ActiveTablesRepository, eventsRepo *repo.EventsRepo) *TableHandler {
	return &TableHandler{
		tablesRepo: activeTablesRepo,
		venuesRepo: venueRepo,
		eventsRepo: eventsRepo,
	}

}

func (h *TableHandler) CodeHandler(w http.ResponseWriter, r *http.Request) {
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
	session, err := h.tablesRepo.GetSessionForTable(code)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		logger.Errorf("error fetching session: %v", err)
		http.Error(w, "Error fetching session", http.StatusInternalServerError)
		return
	}

	if session == nil {
		session = &structs.ActiveTable{
			ClientID:  clientID,
			TableCode: code,
		}
		_, err = h.tablesRepo.TableActive(session)
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

	venue, err := h.venuesRepo.GetVenueByTableCode(r.Context(), code)
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
	menu, err := h.venuesRepo.GetMenuForVenue(r.Context(), venue.ID)
	if err != nil {
		logger.Errorf("error fetching menu: %v", err)
		http.Error(w, "Error fetching menu", http.StatusInternalServerError)
		return
	}
	logger.Infof("found menu: %v", menu)

	menuPage := structs.MenuPage{
		Title:     "Menu",
		Menu:      *menu,
		TableCode: code,
	}
	tmpl := template.Must(template.New("menu.html").Funcs(templateFuncs).ParseFiles("templates/menu.html"))
	err = tmpl.Execute(w, menuPage)
	if err != nil {
		logger.Errorf("error executing template: %v", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

func (h *TableHandler) OrderHandler(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]
	logger.Infof("got code: %v", code)

	session, err := h.tablesRepo.GetSessionForTable(code)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		logger.Errorf("error fetching session: %v", err)
		http.Error(w, "Error fetching session", http.StatusInternalServerError)
		return
	}

	if session == nil {
		logger.Errorf("no active session found for code: %v", code)
		http.Error(w, "No active session found", http.StatusNotFound)
		return
	}

	clientID := ""
	cookie, err := r.Cookie("client_id")
	if err == nil {
		clientID = cookie.Value
	}

	if session.ClientID != clientID {
		tmpl := template.Must(template.New("occupied.html").Funcs(templateFuncs).ParseFiles("templates/occupied.html"))
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		h.orderHandlerPOST(w, r, err, session)
		return
	}

	if r.Method == http.MethodGet {
		var total float64
		for _, item := range session.PreOrder {
			itemTotal := float64(item.Amount) * item.MenuItem.Price
			total += itemTotal
		}

		orderPage := structs.OrderPage{
			Title:        "Current Order",
			Session:      session,
			CurrentTotal: total,
		}
		tmpl := template.Must(template.New("current-order.html").Funcs(templateFuncs).ParseFiles("templates/current-order.html"))
		tmpl.Execute(w, orderPage)
		return
	}
}
func (h *TableHandler) orderHandlerPOST(w http.ResponseWriter, r *http.Request, err error, session *structs.ActiveTable) {
	err = r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	menuItemName := r.FormValue("name")
	if menuItemName == "" {
		logger.Error("name is required")
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	menuItemPriceStr := r.FormValue("price")
	if menuItemPriceStr == "" {
		logger.Error("price is required")
		http.Error(w, "price is required", http.StatusBadRequest)
		return
	}
	menuItemDescStr := r.FormValue("description")

	menuItemPrice, err := strconv.ParseFloat(menuItemPriceStr, 64)
	if err != nil {
		logger.Errorf("error parsing price: %v", err)
		http.Error(w, "Error parsing price", http.StatusBadRequest)
		return
	}
	menuItemAmount, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil {
		logger.Errorf("error parsing amount: %v", err)
		menuItemAmount = 1
	}

	session.PreOrder = append(session.PreOrder, structs.OrderItem{
		MenuItem: &structs.MenuItem{
			Name:        menuItemName,
			Price:       menuItemPrice,
			Description: menuItemDescStr,
		},
		Amount: menuItemAmount,
	})

	// Update the session in the database
	_, err = h.tablesRepo.UpdateSession(session)
	if err != nil {
		logger.Errorf("error updating session: %v", err)
		http.Error(w, "Error updating session", http.StatusInternalServerError)
		return
	}
	return
}

func (h *TableHandler) PlaceOrderHandler(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]
	logger.Infof("got code: %v", code)

	session, err := h.tablesRepo.GetSessionForTable(code)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		logger.Errorf("error fetching session: %v", err)
		http.Error(w, "Error fetching session", http.StatusInternalServerError)
		return
	}

	if session == nil {
		logger.Errorf("no active session found for code: %v", code)
		http.Error(w, "No active session found", http.StatusNotFound)
		return
	}

	clientID := ""
	cookie, err := r.Cookie("client_id")
	if err == nil {
		clientID = cookie.Value
	}

	if session.ClientID != clientID {
		tmpl := template.Must(template.New("occupied.html").Funcs(templateFuncs).ParseFiles("templates/occupied.html"))
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		session.OrderHistory = append(session.OrderHistory, session.PreOrder...)
		session.PreOrder = []structs.OrderItem{}
		_, err = h.tablesRepo.UpdateSession(session)
		if err != nil {
			logger.Errorf("error updating session: %v", err)
			http.Error(w, "Error updating session", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/table/%s", code), http.StatusSeeOther)
		return
	}
}

func (h *TableHandler) OrderHistoryHandler(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]
	logger.Infof("got code: %v", code)

	session, err := h.tablesRepo.GetSessionForTable(code)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		logger.Errorf("error fetching session: %v", err)
		http.Error(w, "Error fetching session", http.StatusInternalServerError)
		return
	}

	if session == nil {
		logger.Errorf("no active session found for code: %v", code)
		http.Error(w, "No active session found", http.StatusNotFound)
		return
	}

	clientID := ""
	cookie, err := r.Cookie("client_id")
	if err == nil {
		clientID = cookie.Value
	}

	if session.ClientID != clientID {
		tmpl := template.Must(template.New("occupied.html").Funcs(templateFuncs).ParseFiles("templates/occupied.html"))
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodGet {
		var total float64
		for _, item := range session.OrderHistory {
			itemTotal := float64(item.Amount) * item.MenuItem.Price
			total += itemTotal
		}

		orderPage := structs.OrderPage{
			Title:        "Order History",
			Session:      session,
			CurrentTotal: total,
		}
		tmpl := template.Must(template.New("order-history.html").Funcs(templateFuncs).ParseFiles("templates/order-history.html"))
		tmpl.Execute(w, orderPage)
	}

	if r.Method == http.MethodPost {
		session.OrderHistory = []structs.OrderItem{}
		_, err = h.tablesRepo.UpdateSession(session)
		if err != nil {
			logger.Errorf("error updating session: %v", err)
			http.Error(w, "Error updating session", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/table/%s", code), http.StatusSeeOther)
		return
	}
}

func (h *TableHandler) CloseOrderHandler(w http.ResponseWriter, r *http.Request) {
	code := mux.Vars(r)["code"]
	logger.Infof("got code: %v", code)
	session, err := h.tablesRepo.GetSessionForTable(code)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		logger.Errorf("error fetching session: %v", err)
		http.Error(w, "Error fetching session", http.StatusInternalServerError)
		return
	}

	if session == nil {
		logger.Errorf("no active session found for code: %v", code)
		http.Error(w, "No active session found", http.StatusNotFound)
		return
	}

	err = r.ParseForm()
	if err != nil {
		logger.Errorf("error parsing form: %v", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	status := r.FormValue("status")
	if status != "paid" && status != "canceled" {
		logger.Errorf("invalid order status: %v", status)
		http.Error(w, "Invalid order status", http.StatusBadRequest)
		return
	}
	logger.Infof("updating session status to: %v", status)
	event := structs.Event{
		Status: status,
		Order:  *session,
	}
	_, err = h.eventsRepo.RecordEvent(&event)
	if err != nil {
		logger.Errorf("error recording event: %v", err)
		http.Error(w, "Error recording event", http.StatusInternalServerError)
		return
	}

	h.tablesRepo.DeleteSession(session.TableCode)

	sessions, err := h.tablesRepo.GetOpenSessions(r.Context())
	if err != nil {
		logger.Errorf("error fetching open sessions: %v", err)
		http.Error(w, "Error fetching open sessions", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.New("open-sessions.html").Funcs(templateFuncs).ParseFiles("templates/open-sessions.html"))
	tmpl.Execute(w, sessions)
}
