package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"tracker/internal/config"
	"tracker/internal/database"
	"tracker/internal/handler"
	"tracker/internal/repository"
)

func main() {
	cfg := config.MustLoad()

	ctx := context.Background()

	db, err := database.New(ctx, cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	defer db.Close()

	repo := repository.NewRegistrationRepository(db)
	taskHandler := handler.NewTaskHandler(repo)

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)                 // Добавляет X-Request-ID
	r.Use(middleware.RealIP)                    // Получает реальный IP клиента
	r.Use(middleware.Logger)                    // Chi-логгер (можно заменить на наш)
	r.Use(middleware.Recoverer)                 // Паник-рекувери
	r.Use(middleware.Timeout(60 * time.Second)) // Таймаут на запрос

	r.Route("/api", func(r chi.Router) {
		r.Get("/tasks", taskHandler.GetTasks)
		r.Post("/tasks", taskHandler.CreateTask)
		r.Get("/tasks/{id}", taskHandler.GetTaskByID)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("starting server on %s (driver=%s)", cfg.Server.Port, cfg.Database.Driver)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server shutdown error: %v", err)
	}

	log.Println("server stopped")
}
