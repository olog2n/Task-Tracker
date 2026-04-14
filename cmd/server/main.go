package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"tracker/internal/config"
	"tracker/internal/database"
	"tracker/internal/handler"
	"tracker/internal/middleware"
	"tracker/internal/repository"

	"github.com/go-chi/chi/v5"
)

func main() {
	ctx := context.Background()

	db, err := database.New(ctx, "file:tracker.db?_foreign_keys=on")
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	apiHandler := &handler.ApiHandler{}

	repo := repository.NewTaskRepository(db)
	taskHandler := handler.NewTaskHandler(repo)

	router := chi.NewRouter()
	router.Use(middleware.Logger)

	router.Get("/api", apiHandler.Version)
	router.Get("/api/health", apiHandler.Health)

	router.Get("/api/tasks", taskHandler.GetTasks)
	router.Get("/api/tasks/{id}", taskHandler.GetTaskByID)
	router.Post("/api/tasks", taskHandler.CreateTask)

	server := &http.Server{
		Addr:    config.DefaultServerPort,
		Handler: router,
	}

	go func() {
		log.Printf(`starting server on localhost%s`, config.DefaultServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}

	log.Println("server stopped")
}
