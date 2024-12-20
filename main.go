package main

import (
	"context"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
	"vortex.studio/account/internal/handlers"
	"vortex.studio/account/internal/repo"
)

func main() {
	router := mux.NewRouter()

	clientOptions := options.Client().ApplyURI(os.Getenv("MONGODB_CONN_STRING"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			log.Fatalf("Failed to disconnect MongoDB: %v", err)
		}
	}()

	db := client.Database("account")
	venueRepository := repo.NewVenueRepository(db)
	activeTablesRepo := repo.NewActiveTablesRepository(db)
	eventsRepo := repo.NewEventsRepo(db)
	menuRepo := repo.NewMenuRepository(db)

	adminHandler := handlers.NewAdminHandler(*venueRepository, activeTablesRepo, menuRepo)
	tablesHandler := handlers.NewTablesHandler(venueRepository, activeTablesRepo, eventsRepo)

	router.HandleFunc("/admin", adminHandler.AccountHandler).Methods("GET")
	router.HandleFunc("/table", adminHandler.AddTableHandler).Methods("POST")
	router.HandleFunc("/login", adminHandler.LoginHandler).Methods("POST")
	router.HandleFunc("/logout", adminHandler.LogoutHandler).Methods("GET")
	router.HandleFunc("/venue", adminHandler.VenueHandler).Methods("POST")

	router.HandleFunc("/table/{code}", tablesHandler.CodeHandler).Methods("GET", "POST")
	router.HandleFunc("/order/{code}", tablesHandler.OrderHandler).Methods("POST", "GET")
	router.HandleFunc("/order/{code}/place", tablesHandler.PlaceOrderHandler).Methods("POST")
	router.HandleFunc("/history/{code}", tablesHandler.OrderHistoryHandler).Methods("GET")
	router.HandleFunc("/close/{code}", tablesHandler.CloseOrderHandler).Methods("POST")

	log.Fatal(http.ListenAndServe(":9090", router))
}
