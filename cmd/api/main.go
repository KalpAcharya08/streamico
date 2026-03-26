package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"streaming-backend/internal/auth"
	"streaming-backend/internal/common/cache"
	"streaming-backend/internal/common/config"
	"streaming-backend/internal/common/database"
	"streaming-backend/internal/common/logger"
	"streaming-backend/internal/common/middleware"
	"streaming-backend/internal/user"
	"streaming-backend/internal/video"
)

// =============================================================================
// 📖 LEARNING NOTES — Step 3: MySQL Database Connection
// =============================================================================
//
// What changed from Step 2:
//
//   BEFORE (Step 2)                    →  AFTER (Step 3)
//   ──────────────────────────────────────────────────────────────────────────
//   No database                        →  MySQL via database/sql
//   Config: Port, Env only             →  Config: + DB settings (host, port, etc.)
//   Health: just "ok"                  →  Health: checks DB connectivity too
//
// New concepts:
//   - database/sql  → Go's standard database interface (works with any DB)
//   - sql.DB        → connection pool (built-in, unlike PostgreSQL's pgxpool)
//   - db.Ping(ctx)  → verifies the database is reachable
//   - db.Close()    → closes all connections (called on shutdown)
//   - Docker Compose → runs MySQL in a container with one command
// =============================================================================

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.New(cfg.Env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// ─── Step C: Connect to MySQL ────────────────────────────────────────
	db, err := database.NewMySQLPool(context.Background(), database.MySQLConfig{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		DBName:   cfg.DB.Name,
	}, log)
	if err != nil {
		log.Fatal("❌ failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// ─── Step C2: Connect to Redis ───────────────────────────────────────
	redisClient, err := cache.NewRedisClient(context.Background(), cache.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
	}, log)
	if err != nil {
		log.Fatal("❌ failed to connect to redis", zap.Error(err))
	}
	defer redisClient.Close()

	// ─── Step D: Create router & middleware ───────────────────────────────
	r := chi.NewRouter()
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger(log))

	// ─── Step E: Initialize Services (Dependency Injection) ────────────────
	
	// Create caches
	authCache := auth.NewCache(redisClient)

	// Build the Auth chain: Repository -> Service -> Handler
	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, cfg.JWT.Secret, cfg.JWT.TTL)
	authHandler := auth.NewHandler(authService, authCache)

	// Auth Middleware
	authMiddle := auth.NewMiddleware(cfg.JWT.Secret, authCache)

	// Build the User chain: Repository -> Service -> Handler
	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// Build the Video chain: Repository -> Service -> Handler
	videoCache := video.NewCache(redisClient)
	videoRepo := video.NewRepository(db)
	videoService := video.NewService(videoRepo, videoCache, cfg.PexelsAPIKey)
	videoHandler := video.NewHandler(videoService)

	// ─── Step F: Register routes ─────────────────────────────────────────
	r.Get("/health", handleHealth(db))
	r.Get("/", handleWelcome)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handleHealth(db))

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddle.RequireAuth)
			
			r.Post("/auth/logout", authHandler.Logout)
			
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", userHandler.GetMe)
			})
			r.Route("/videos", func(r chi.Router) {
				r.Post("/", videoHandler.UploadVideo)
				r.Get("/", videoHandler.ListUserVideos)
				r.Get("/explore", videoHandler.ExploreVideos)
				r.Get("/{id}", videoHandler.GetVideo)
			})
		})
	})

	// Serve uploaded files statically
	os.MkdirAll("uploads/videos", 0755)
	r.Get("/uploads/videos/*", func(w http.ResponseWriter, r *http.Request) {
		http.StripPrefix("/uploads/videos/", http.FileServer(http.Dir("./uploads/videos"))).ServeHTTP(w, r)
	})

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("🚀 server starting",
			zap.String("port", cfg.Port),
			zap.String("env", cfg.Env),
			zap.String("health", fmt.Sprintf("http://localhost:%s/health", cfg.Port)),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("❌ server failed to start", zap.Error(err))
		}
	}()

	// ─── Step G: Graceful shutdown ───────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	log.Info("⚠️  shutdown signal received", zap.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("❌ server forced to shutdown", zap.Error(err))
	}

	log.Info("✅ server stopped gracefully")
}

// =============================================================================
// HANDLERS
// =============================================================================

// handleHealth checks database connectivity and returns server status.
//
// 📖 PATTERN: We use an interface instead of a concrete type (*sql.DB).
// This means we only require a "Pingable" thing — easier to test with mocks.
type Pinger interface {
	PingContext(ctx context.Context) error
}

func handleHealth(db Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbStatus := "connected"
		if err := db.PingContext(r.Context()); err != nil {
			dbStatus = fmt.Sprintf("error: %v", err)
		}

		response := map[string]interface{}{
			"status":    "ok",
			"message":   "Streaming backend is running 🎬",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "0.5.0",
			"database":  dbStatus,
		}
		writeJSON(w, http.StatusOK, response)
	}
}

func handleWelcome(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"name":    "Streaming Backend API",
		"version": "0.5.0",
		"endpoints": map[string]string{
			"health":     "/health",
			"api_health": "/api/v1/health",
		},
	}
	writeJSON(w, http.StatusOK, response)
}

// =============================================================================
// HELPERS
// =============================================================================

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonBytes)
}
