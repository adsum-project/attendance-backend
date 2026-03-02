package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/adsum-project/attendance-backend/internal/db"
	authhandlers "github.com/adsum-project/attendance-backend/internal/handlers/auth"
	timetablehandlers "github.com/adsum-project/attendance-backend/internal/handlers/timetable"
	userhandlers "github.com/adsum-project/attendance-backend/internal/handlers/user"
	verificationhandlers "github.com/adsum-project/attendance-backend/internal/handlers/verification"
	"github.com/adsum-project/attendance-backend/internal/middleware"
	authrepo "github.com/adsum-project/attendance-backend/internal/repositories/auth"
	timetablerepo "github.com/adsum-project/attendance-backend/internal/repositories/timetable"
	verificationrepo "github.com/adsum-project/attendance-backend/internal/repositories/verification"
	"github.com/adsum-project/attendance-backend/internal/services/auth"
	"github.com/adsum-project/attendance-backend/internal/services/graph"
	"github.com/adsum-project/attendance-backend/internal/services/timetable"
	"github.com/adsum-project/attendance-backend/internal/services/verification"
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
	authService, err := auth.NewAuthService(sessionRepo)
	if err != nil {
		log.Fatal("Error initializing auth: ", err)
	}
	authProvider, err := authhandlers.NewAuthProvider(authService)
	if err != nil {
		log.Fatal("Error initializing auth provider: ", err)
	}

	graphService, err := graph.NewGraphService()
	if err != nil {
		log.Fatal("Error initializing Graph service: ", err)
	}

	timetableRepo := timetablerepo.NewTimetableRepository(dbProvider.DB)
	timetableService, err := timetable.NewTimetableService(timetableRepo, graphService)
	if err != nil {
		log.Fatal("Error initializing timetable service: ", err)
	}
	timetableProvider, err := timetablehandlers.NewTimetableProvider(timetableService)
	if err != nil {
		log.Fatal("Error initializing timetable provider: ", err)
	}

	verificationRepo := verificationrepo.NewVerificationRepository(dbProvider.DB)
	verificationService, err := verification.NewVerificationService(verificationRepo, timetableService)
	if err != nil {
		log.Fatal("Error initializing verification service: ", err)
	}

	go verificationService.RunQRTokenCleanup(5 * time.Minute)
	go verificationService.RunAbsentRecordProcessor(5 * time.Minute)
	verificationProvider, err := verificationhandlers.NewVerificationProvider(verificationService)
	if err != nil {
		log.Fatal("Error initializing verification provider: ", err)
	}
	userProvider, err := userhandlers.NewUserProvider(graphService)
	if err != nil {
		log.Fatal("Error initializing user provider: ", err)
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
		r.Get("/records", verificationProvider.GetRecords).Use(middleware.RequireAuth(authService, "default", "admin", "staff"))
		r.Patch("/records/{record_id}", verificationProvider.PatchRecordStatus).Use(middleware.RequireAuth(authService, "admin", "staff"))

		r.Get("/qr/verify", verificationProvider.QRVerify).Use(middleware.RequireAuthWithRedirect(authService, os.Getenv("FRONTEND_URL")+"/", "default"))
		r.Get("/qr", verificationProvider.QRStream).Use(middleware.RequireAuth(authService, "admin"))
	})

	r.Group("/v1/timetable", func() {
		r.Get("/me", timetableProvider.GetOwnTimetable).Use(middleware.RequireAuth(authService, "default"))
		r.Get("/node", timetableProvider.GetNodeTimetable).Use(middleware.RequireAuth(authService, "admin"))
		r.Get("/node/room", timetableProvider.GetNodeRoom).Use(middleware.RequireAuth(authService, "admin"))
		r.Put("/node/room", timetableProvider.UpdateNodeRoom).Use(middleware.RequireAuth(authService, "admin"))
		r.Delete("/node/room", timetableProvider.DeleteNodeRoom).Use(middleware.RequireAuth(authService, "admin"))

		r.Get("/courses", timetableProvider.GetCourses).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Get("/courses/me", timetableProvider.GetOwnCourses).Use(middleware.RequireAuth(authService, "default"))
		r.Get("/courses/{course_id}", timetableProvider.GetCourse).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Post("/courses", timetableProvider.CreateCourse).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Patch("/courses/{course_id}", timetableProvider.UpdateCourse).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Delete("/courses/{course_id}", timetableProvider.DeleteCourse).Use(middleware.RequireAuth(authService, "admin", "staff"))

		r.Get("/courses/{course_id}/modules", timetableProvider.GetCourseModules).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Post("/courses/{course_id}/modules/{module_id}", timetableProvider.AssignModuleToCourse).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Delete("/courses/{course_id}/modules/{module_id}", timetableProvider.UnassignModuleFromCourse).Use(middleware.RequireAuth(authService, "admin", "staff"))

		r.Get("/courses/{course_id}/students", timetableProvider.GetCourseStudents).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Post("/courses/{course_id}/students/{user_id}", timetableProvider.AssignStudentToCourse).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Delete("/courses/{course_id}/students/{user_id}", timetableProvider.UnassignStudentFromCourse).Use(middleware.RequireAuth(authService, "admin", "staff"))

		r.Get("/modules", timetableProvider.GetModules).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Get("/modules/{module_id}", timetableProvider.GetModule).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Post("/modules", timetableProvider.CreateModule).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Patch("/modules/{module_id}", timetableProvider.UpdateModule).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Delete("/modules/{module_id}", timetableProvider.DeleteModule).Use(middleware.RequireAuth(authService, "admin", "staff"))

		r.Get("/modules/{module_id}/courses", timetableProvider.GetModuleCourses).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Get("/modules/{module_id}/classes", timetableProvider.GetClasses).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Get("/modules/{module_id}/classes/{class_id}", timetableProvider.GetClass).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Post("/modules/{module_id}/classes", timetableProvider.CreateClass).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Patch("/modules/{module_id}/classes/{class_id}", timetableProvider.UpdateClass).Use(middleware.RequireAuth(authService, "admin", "staff"))
		r.Delete("/modules/{module_id}/classes/{class_id}", timetableProvider.DeleteClass).Use(middleware.RequireAuth(authService, "admin", "staff"))
	})

	r.Group("/v1/users", func() {
		r.Get("", userProvider.GetUsers).Use(middleware.RequireAuth(authService, "admin", "staff"))
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.OK(w, "API is running", map[string]string{"since": serverStartTime})
	})

	r.StartServer(":8080")
}
