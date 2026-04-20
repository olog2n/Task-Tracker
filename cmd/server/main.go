package main

import (
	"context"
	"database/sql"
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
	"tracker/internal/tracemiddleware"
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
	cfg := loadConfig()

	db := initDatabase(cfg)
	defer db.Close()

	jwtService := initJWTService(cfg)

	repos := initRepositories(db)

	metadata := initMetadataService()

	handlers := initHandlers(
		cfg,
		db,
		repos,
		jwtService,
		metadata,
	)

	router := initRouter(handlers, jwtService, repos)

	runServer(cfg, router)
}

func loadConfig() *config.Config {
	log.Println("Loading configuration...")
	cfg := config.MustLoad()
	log.Println("Configuration loaded successfully")
	return cfg
}

func initDatabase(cfg *config.Config) *sql.DB {
	log.Println("Initializing database...")
	db, err := database.New(context.Background(), cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database service UP")
	return db
}

func initJWTService(cfg *config.Config) *auth.JWTService {
	log.Println("Initializing JWT service...")
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
		log.Fatalf("Failed to initialize JWT service: %v", err)
	}
	log.Println("JWT Service UP")

	return jwtService
}

func initMetadataService() *handler.MetadataService {
	log.Println("Initializing Metadata service...")
	return handler.NewMetadataService(
		"0.2.0",
		time.Now().Format(time.RFC822Z),
	)
}

type Repositories struct {
	User    repository.UserRepository
	Task    repository.TaskRepository
	Project repository.ProjectRepository
}

func initRepositories(db *sql.DB) *Repositories {
	log.Println("Initializing repositories...")
	return &Repositories{
		User:    repository.NewUserRepository(db),
		Task:    repository.NewTaskRepository(db),
		Project: repository.NewProjectRepository(db),
	}
}

type Handlers struct {
	Auth    *handler.AuthHandler
	Task    *handler.TaskHandler
	Version *handler.VersionHandler
	Health  *handler.HealthHandler
	User    *handler.UserHandler
	Project *handler.ProjectHandler
}

func initHandlers(
	cfg *config.Config,
	db *sql.DB,
	repos *Repositories,
	jwtService *auth.JWTService,
	metadata *handler.MetadataService,
) *Handlers {
	log.Println("Initializing handlers...")
	return &Handlers{
		Auth: handler.NewAuthHandler(
			repos.User,
			jwtService,
			cfg.Auth.CookieSecure,
			cfg.Auth.CookieDomain,
			cfg.Auth.CookiePath,
			cfg.Auth.CookieNameAccess,
			cfg.Auth.CookieNameRefresh,
		),
		Task:    handler.NewTaskHandler(repos.Task),
		Version: handler.NewVersionHandler(metadata),
		Health:  handler.NewHealthHandler(metadata, db),
		User:    handler.NewUserHandler(repos.User),
		Project: handler.NewProjectHandler(repos.Project, repos.User),
	}
}

func initRouter(handlers *Handlers, jwtService *auth.JWTService, repos *Repositories) *chi.Mux {
	log.Println("Setting up router...")

	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(tracemiddleware.Logger)

	// Swagger UI
	docs.SwaggerInfo.Title = "Issue Tracker API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Description = "Simple and fast task tracker with jwt auth"
	docs.SwaggerInfo.Host = "localhost:6969"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	//Health checks
	r.Get("/health", handlers.Health.Live)
	r.Get("/ready", handlers.Health.Ready)
	r.Get("/version", handlers.Version.ServeHTTP)

	// Public routes (authentication)
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", handlers.Auth.Register)
		r.Post("/login", handlers.Auth.Login)
		r.Post("/refresh", handlers.Auth.RefreshToken)
	})

	// Protected routes
	r.Route("/api", func(r chi.Router) {
		r.Use(tracemiddleware.AuthMiddleware(jwtService))

		r.Route("/projects", func(r chi.Router) {
			r.Post("/", handlers.Project.CreateProject)
			r.Get("/", handlers.Project.GetProjects)

			r.Route("/{project_id}", func(r chi.Router) {
				r.Use(tracemiddleware.ProjectAuthMiddleware(repos.Project))

				r.Get("/", handlers.Project.GetProject)
				r.Put("/", handlers.Project.UpdateProject)
				r.Delete("/", handlers.Project.DeleteProject)

				r.Route("/tasks", func(r chi.Router) {
					r.Get("/", handlers.Task.GetTasks)
					r.Post("/", handlers.Task.CreateTask)
					r.Get("/{task_id}", handlers.Task.GetTaskByID)
					r.Put("/{task_id}", handlers.Task.UpdateTask)
					r.Delete("/{task_id}", handlers.Task.DeleteTask)
				})

				r.Route("/members", func(r chi.Router) {
					r.Get("/", handlers.Project.GetMembers)
					r.Post("/", handlers.Project.AddMember)
					r.Put("/{user_id}", handlers.Project.UpdateMemberRole)
					r.Delete("/{user_id}", handlers.Project.RemoveMember)
				})
			})
		})

		r.Route("/users", func(r chi.Router) {
			r.Delete("/{id}", handlers.User.DeactivateUser)
			r.Post("/{id}/reactivate", handlers.User.ReactivateUser)
		})

		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", handlers.Task.GetTasks)
			r.Post("/", handlers.Task.CreateTask)
			r.Get("/{id}", handlers.Task.GetTaskByID)
			r.Put("/{id}", handlers.Task.UpdateTask)
			r.Delete("/{id}", handlers.Task.DeleteTask)
		})
	})

	log.Println("Router UP")
	return r
}

func runServer(cfg *config.Config, handler http.Handler) {
	log.Println("Starting server...")

	server := &http.Server{
		Addr:         cfg.Server.Port,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("Server listening on %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped gracefully")
}
