package main

import (
	"context"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"vortex.studio/account/internal/handlers"
	"vortex.studio/account/internal/repo"
)

func main() {
	router := mux.NewRouter()

	// TODO: Get mongoDB connection string from environment variable or config file
	clientOptions := options.Client().ApplyURI("mongodb://admin:secretpassword@localhost:27017")
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

	handlers := handlers.NewHandler(*venueRepository, activeTablesRepo)

	router.HandleFunc("/admin", handlers.AccountHandler).Methods("GET")
	router.HandleFunc("/add-table", handlers.AddTableHandler).Methods("POST")
	router.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	router.HandleFunc("/logout", handlers.LogoutHandler).Methods("GET")
	router.HandleFunc("/venue", handlers.VenueHandler).Methods("POST")
	router.HandleFunc("/table/{code}", handlers.CodeHandler).Methods("GET", "POST")

	log.Fatal(http.ListenAndServe(":9090", router))
}
