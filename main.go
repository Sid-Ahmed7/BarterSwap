package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("database open error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("database connection error: %v", err)
	}
	log.Println("connected to database")

	store := &DB{db}
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/users", handleCreateUser(store))
	mux.HandleFunc("GET /api/users/{id}", handleGetUser(store))
	mux.HandleFunc("PUT /api/users/{id}", handleUpdateUser(store))
	mux.HandleFunc("GET /api/users/{id}/skills", handleGetUserSkills(store))
	mux.HandleFunc("PUT /api/users/{id}/skills", handleSetUserSkills(store))

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server started on port %s", port)
	if err := http.ListenAndServe(":"+port, buildHandler(mux)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
