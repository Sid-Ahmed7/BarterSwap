package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "barterswap/docs"
	"barterswap/internal/handler"
	"barterswap/internal/middleware"
	"barterswap/internal/store"

	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title BarterSwap API
// @version 1.0
// @description API REST d'échange de compétences via crédits-temps.
// @host localhost:8080
// @BasePath /
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

	store := &store.DB{DB: db}
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/users", handler.HandleCreateUser(store))
	mux.HandleFunc("GET /api/users/{id}", handler.HandleGetUser(store))
	mux.HandleFunc("PUT /api/users/{id}", handler.HandleUpdateUser(store))
	mux.HandleFunc("GET /api/users/{id}/skills", handler.HandleGetUserSkills(store))
	mux.HandleFunc("PUT /api/users/{id}/skills", handler.HandleSetUserSkills(store))

	mux.HandleFunc("POST /api/services", handler.HandleCreateService(store))
	mux.HandleFunc("GET /api/services/{id}", handler.HandleGetService(store))
	mux.HandleFunc("GET /api/services", handler.HandleListServices(store))
	mux.HandleFunc("PUT /api/services/{id}", handler.HandleUpdateService(store))
	mux.HandleFunc("DELETE /api/services/{id}", handler.HandleDeleteService(store, store))

	mux.HandleFunc("POST /api/exchanges", handler.HandleCreateExchange(store, store, store))
	mux.HandleFunc("GET /api/exchanges", handler.HandleListExchanges(store))
	mux.HandleFunc("GET /api/exchanges/{id}", handler.HandleGetExchange(store))
	mux.HandleFunc("PUT /api/exchanges/{id}/accept", handler.HandleAcceptExchange(store))
	mux.HandleFunc("PUT /api/exchanges/{id}/reject", handler.HandleRejectExchange(store))
	mux.HandleFunc("PUT /api/exchanges/{id}/complete", handler.HandleCompleteExchange(store))
	mux.HandleFunc("PUT /api/exchanges/{id}/cancel", handler.HandleCancelExchange(store))

	mux.HandleFunc("POST /api/exchanges/{id}/review", handler.HandleCreateReview(store, store))
	mux.HandleFunc("GET /api/users/{id}/reviews", handler.HandleGetUserReviews(store))
	mux.HandleFunc("GET /api/services/{id}/reviews", handler.HandleGetServiceReviews(store))
	mux.HandleFunc("GET /api/users/{id}/stats", handler.HandleGetUserStats(store))
	mux.HandleFunc("GET /swagger/{any...}", httpSwagger.WrapHandler)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server started on port %s", port)
	if err := http.ListenAndServe(":"+port, middleware.BuildHandler(mux)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
