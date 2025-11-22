package main

import (
	"context"
	"fmt"
	"github.com/mytheresa/go-hiring-challenge/app/categories"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mytheresa/go-hiring-challenge/app/catalog"
	"github.com/mytheresa/go-hiring-challenge/app/database"
	"github.com/mytheresa/go-hiring-challenge/models"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	// signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize database connection
	db, close := database.New(
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)
	defer close()

	// Initialize handlers
	prodRepo := models.NewProductsRepository(db)
	catRepo := models.NewCategoriesRepository(db)
	catalogHandler := catalog.NewCatalogHandler(prodRepo)
	categoriesHandler := categories.NewCategoriesHandler(catRepo)

	// Set up routing
	mux := http.NewServeMux()
	mux.HandleFunc("GET /catalog", catalogHandler.HandleGet)
	mux.HandleFunc("GET /catalog/{code}", catalogHandler.HandleGetDetail)
	mux.HandleFunc("GET /categories", categoriesHandler.HandleGet)
	mux.HandleFunc("POST /categories", categoriesHandler.HandleCreate)

	// Set up the HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf("localhost:%s", os.Getenv("HTTP_PORT")),
		Handler: mux,
	}

	// Start the server
	go func() {
		log.Printf("Starting server on http://%s", srv.Addr)
		log.Printf("Available endpoints:")
		log.Printf("  GET  /catalog, for fetching products filtered by category and price_less_than, paginated by offset and limit")
		log.Printf("  GET  /catalog/{code}, for fetching products details with code")
		log.Printf("  GET  /categories, for fetching all categories")
		log.Printf("  POST /categories, for creating a new category")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %s", err)
		}

		log.Println("Server stopped gracefully")
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")
	srv.Shutdown(ctx)
	stop()
}
