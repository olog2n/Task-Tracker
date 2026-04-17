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
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"tracker/docs"
	"tracker/internal/auth"
	"tracker/internal/config"
	"tracker/internal/database"
	"tracker/internal/handler"
	"tracker/internal/repository"
	"tracker/internal/traceMiddleware"
)

// @title           Issue Tracker API
// @version         0.1.0
// @description     Простой и быстрый трекер задач с JWT аутентификацией
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name   GPLv3
// @license.url    https://opensource.org/license/gpl-3.0

// @host      localhost:6969
// @BasePath  /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Введите токен в формате: "Bearer <token>"

func main() {
	cfg := config.MustLoad()

	ctx := context.Background()

	db, err := database.New(ctx, cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed to init database: %v", err)
	}
	log.Printf("database connection OK")
	defer db.Close()

	taskRepo := repository.NewTaskRepository(db)
	userRepo := repository.NewUserRepository(db)

	metadata := handler.NewMetadataService("0.1.0", time.Now().Format(time.RFC822Z))
	jwtService, err := auth.NewJWTService(
		cfg.Auth.JWTAlgorithm,
		cfg.Auth.JWTSecret,
		cfg.Auth.JWTPrivateKey,
		cfg.Auth.JWTPublicKey,
		cfg.Auth.JWTKeyID,
		cfg.Auth.JWTExpiry,
		cfg.Auth.JWTRefreshExpiry,
	)

	if err != nil {
		log.Fatalf("JWT service failed to init: %v", err)
	}

	// testClaims, err := jwtService.ValidateToken()
	// if err != nil {
	// 	log.Fatalf("JWT validation failed: %v", err)
	// }
	// log.Printf("JWT service OK")

	healthHandler := handler.NewHealthHandler(metadata, db)
	versionHandler := handler.NewVersionHandler(metadata)

	taskHandler := handler.NewTaskHandler(taskRepo)
	authHandler := handler.NewAuthHandler(
		userRepo,
		jwtService,
		cfg.Auth.CookieSecure,
		cfg.Auth.CookieDomain,
		cfg.Auth.CookiePath,
		cfg.Auth.CookieNameAccess,
		cfg.Auth.CookieNameRefresh,
	)

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(traceMiddleware.Logger)

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		// r.Post("/logout", authHandler.Logout)
		r.Post("/refresh", authHandler.RefreshToken)
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(traceMiddleware.AuthMiddleware(jwtService))

		r.Route("/tasks", func(r chi.Router) {
			r.Post("/", taskHandler.CreateTask)
			r.Get("/", taskHandler.GetTasks)
			r.Get("/{id}", taskHandler.GetTaskByID)
			r.Put("/{id}", taskHandler.UpdateTask)
			r.Delete("/{id}", taskHandler.DeleteTask)
		})
	})

	r.Get("/health", healthHandler.Live)
	r.Get("/ready", healthHandler.Ready)
	r.Get("/version", versionHandler.ServeHTTP)

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	docs.SwaggerInfo.Title = "Issue Tracker API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Description = "Simple and fast task tracker with jwt auth"
	docs.SwaggerInfo.Host = "localhost:6969"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

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
