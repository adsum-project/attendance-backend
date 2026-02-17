package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/adsum-project/attendance-backend/internal/db"
	authhandlers "github.com/adsum-project/attendance-backend/internal/handlers/auth"
	verificationhandlers "github.com/adsum-project/attendance-backend/internal/handlers/verification"
	"github.com/adsum-project/attendance-backend/internal/middleware"
	authrepo "github.com/adsum-project/attendance-backend/internal/repo/auth"
	"github.com/adsum-project/attendance-backend/internal/services/auth"
	"github.com/adsum-project/attendance-backend/pkg/router"
	"github.com/adsum-project/attendance-backend/pkg/utils"
	"github.com/adsum-project/attendance-backend/pkg/utils/response"
	"github.com/joho/godotenv"
)

func main() {
	r := router.NewRouter()

	if utils.GetEnvironment() != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	getCORSOrigins := func() []string {
		origins := os.Getenv("CORS_ORIGINS")
		if origins == "" {
			log.Fatal("CORS_ORIGINS environment variable is required")
		}
		return strings.Split(origins, ",")
	}

	corsMiddleware := middleware.NewCORS(middleware.CORSOptions{
		AllowedOrigins:   getCORSOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
		ExposedHeaders:   []string{},
		AllowCredentials: true,
		MaxAge:           3600,
	})

	r.Use(corsMiddleware)
	r.Use(middleware.NewRequestLogger())

	dbConn, err := db.OpenFromEnv()
	if err != nil {
		log.Fatal("Error initializing db: ", err)
	}
	defer dbConn.Close()

	dbProvider := db.NewDbProvider(dbConn)

	sessionRepo := authrepo.NewSessionRepository(
		dbProvider.DB,
		time.Duration(auth.DefaultCookieMaxAge)*time.Second,
	)
	authService, err := auth.NewAuth(sessionRepo)
	if err != nil {
		log.Fatal("Error initializing auth: ", err)
	}
	authProvider, err := authhandlers.NewAuthProvider(authService)
	if err != nil {
		log.Fatal("Error initializing auth provider: ", err)
	}

	verificationProvider, err := verificationhandlers.NewVerificationProvider()
	if err != nil {
		log.Fatal("Error initializing verification provider: ", err)
	}

	serverStartTime := time.Now().Format(time.RFC3339)

	r.Group("/v1/auth", func() {
		r.Get("/login", authProvider.Login).Use(middleware.RequireNoAuth(authService))
		r.Get("/callback", authProvider.Callback).Use(middleware.RequireNoAuth(authService))
		r.Post("/logout", authProvider.Logout).Use(middleware.RequireAuth(authService))
		r.Get("/me", authProvider.Me).Use(middleware.RequireAuth(authService))
	})

	r.Group("/v1/verification", func() {
		r.Get("/embeddings", verificationProvider.GetEmbedding).Use(middleware.RequireAuth(authService))
		r.Post("/embeddings", verificationProvider.CreateEmbedding).Use(middleware.RequireAuth(authService))
		r.Put("/embeddings", verificationProvider.UpdateEmbedding).Use(middleware.RequireAuth(authService))
		r.Delete("/embeddings", verificationProvider.DeleteEmbedding).Use(middleware.RequireAuth(authService))
		r.Post("/embeddings/verify", verificationProvider.VerifyEmbedding).Use(middleware.RequireAuth(authService))
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, "API is running", map[string]string{"since": serverStartTime})
	})

	r.StartServer(":8080")
}
