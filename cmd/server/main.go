package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/mytheresa/go-hiring-challenge/app/catalog"
	"github.com/mytheresa/go-hiring-challenge/app/category"
	"github.com/mytheresa/go-hiring-challenge/app/database"
	categoryRepo "github.com/mytheresa/go-hiring-challenge/app/repos/category"
	"github.com/mytheresa/go-hiring-challenge/app/repos/products"
)

func main() {
	srv, closeDBCon := initServer()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	go serve(srv)

	// Wait for interrupt signal to gracefully shutdown the server
	<-ctx.Done()
	log.Println("Shutting down server...")
	srv.Shutdown(ctx)
	closeDBCon()
	stop()
}

func initServer() (*http.Server, func() error) {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	db, closeDBCon := database.New(
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"),
	)

	productsRepo := products.NewGormRepo(db)
	categoriesRepo := categoryRepo.NewGormRepo(db)

	catalogH := catalog.NewCatalogHandler(productsRepo)
	categoryH := category.NewCategoryHandler(categoriesRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /catalog", catalogH.HandleGet)
	mux.HandleFunc("GET /catalog/{code}", catalogH.HandleGetSpecific)
	mux.HandleFunc("GET /categories", categoryH.HandleGet)
	mux.HandleFunc("POST /categories", categoryH.HandlePost)

	return &http.Server{
		Addr:    fmt.Sprintf("localhost:%s", os.Getenv("HTTP_PORT")),
		Handler: mux,
	}, closeDBCon
}

func serve(srv *http.Server) {
	log.Printf("Starting server on http://%s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %s", err)
	}

	log.Println("Server stopped gracefully")
}
