package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/encertia/backend/internal/auth"
	"github.com/encertia/backend/internal/course"
	"github.com/encertia/backend/internal/db"
	"github.com/encertia/backend/internal/evaluation"
	"github.com/encertia/backend/internal/match"
	"github.com/encertia/backend/internal/quiz"
	"github.com/encertia/backend/internal/shared"
	"github.com/encertia/backend/internal/user"
	"github.com/gin-gonic/gin"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	log.Println("Iniciant Encertia Backend API...")

	// 1. Environment & DB configuration
	dbCfg := db.Config{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "5432"),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", "postgres"),
		DBName:          getEnv("DB_NAME", "encertia"),
		SSLMode:         getEnv("DB_SSLMODE", "disable"),
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
	}

	jwtSecret := getEnv("JWT_SECRET", "encertia-super-secret-key-change-in-production-min-32-chars")
	serverPort := getEnv("PORT", "8080")
	autoMigrate := getEnv("AUTO_MIGRATE", "true")

	// Storage configuration
	storageCfg := shared.StorageConfig{
		R2AccountID:       getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:      getEnv("R2_BUCKET_NAME", ""),
		R2PublicURL:       getEnv("R2_PUBLIC_URL", ""),
		LocalUploadDir:    getEnv("UPLOAD_DIR", "./uploads"),
		BaseURL:           getEnv("BASE_URL", ""),
	}

	// 2. Connect to Database
	dbConn, err := db.Connect(dbCfg)
	if err != nil {
		log.Printf("[Avís] No s'ha pogut connectar a la base de dades: %v", err)
		log.Println("[Avís] El servidor continuarà arrencant (assegura't que PostgreSQL estigui actiu).")
	} else {
		defer dbConn.Close()
		log.Println("[DB] Connexió amb PostgreSQL establerta correctament.")

		// 3. Run auto-migrations if enabled
		if autoMigrate == "true" {
			if err := db.RunMigrations(dbConn); err != nil {
				log.Printf("[DB Error] Error aplicant migracions: %v", err)
			}
		}
	}

	// 4. Initialize Domain Modules & Services
	storageSvc := shared.NewStorageService(storageCfg)

	// Auth Domain
	authRepo := auth.NewRepository(dbConn)
	authSvc := auth.NewService(authRepo, auth.Config{
		JWTSecret:            jwtSecret,
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "encertia",
	})
	authHandler := auth.NewHandler(authSvc)

	// User Domain
	userRepo := user.NewRepository(dbConn)
	userSvc := user.NewService(userRepo)
	userHandler := user.NewHandler(userSvc)

	// Quiz Domain
	quizRepo := quiz.NewRepository(dbConn)
	quizSvc := quiz.NewService(quizRepo)
	quizHandler := quiz.NewHandler(quizSvc, storageSvc)

	appBaseURL := getEnv("APP_BASE_URL", getEnv("FRONTEND_URL", "http://localhost:5173"))

	// Match Domain
	matchHub := match.NewHub()
	matchRepo := match.NewRepository(dbConn)
	matchSvc := match.NewService(matchRepo, matchHub, appBaseURL)
	matchHandler := match.NewHandler(matchSvc, authHandler.TokenValidatorAdapter(), matchHub)

	// Evaluation Domain
	evalRepo := evaluation.NewRepository(dbConn)
	evalSvc := evaluation.NewService(evalRepo, quizSvc)
	evalHandler := evaluation.NewHandler(evalSvc)
	matchSvc.RegisterFinishedListener(evalSvc)

	// Course Domain
	courseRepo := course.NewRepository(dbConn)
	courseSvc := course.NewService(courseRepo)
	courseHandler := course.NewHandler(courseSvc)

	// Middleware
	authMiddleware := shared.AuthMiddleware(authHandler.TokenValidatorAdapter())

	// 5. Setup Gin Router
	ginMode := getEnv("GIN_MODE", gin.ReleaseMode)
	gin.SetMode(ginMode)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), shared.CORSMiddleware())

	// Static route for local uploads
	router.Static("/uploads", storageCfg.LocalUploadDir)

	// Health Check
	healthHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "encertia-backend",
		})
	}
	router.Match([]string{"GET", "HEAD"}, "/", healthHandler)
	router.Match([]string{"GET", "HEAD"}, "/health", healthHandler)
	router.Match([]string{"GET", "HEAD"}, "/api/health", healthHandler)

	// Register routes directly at root (/auth/*, /users/*, /quizzes/*, /matches/*, /uploads/*, /ws/match/*, /courses/*) as per OpenAPI contract
	rootGroup := router.Group("")
	authHandler.RegisterRoutes(rootGroup, authMiddleware)
	userHandler.RegisterRoutes(rootGroup, authMiddleware)
	quizHandler.RegisterRoutes(rootGroup, authMiddleware)
	matchHandler.RegisterRoutes(rootGroup, authMiddleware)
	evalHandler.RegisterRoutes(rootGroup, authMiddleware)
	courseHandler.RegisterRoutes(rootGroup, authMiddleware)

	// Also support /api prefix for proxy convenience (/api/auth/*, /api/users/*, /api/quizzes/*, /api/matches/*, /api/uploads/*, /api/ws/match/*, /api/courses/*)
	apiGroup := router.Group("/api")
	authHandler.RegisterRoutes(apiGroup, authMiddleware)
	userHandler.RegisterRoutes(apiGroup, authMiddleware)
	quizHandler.RegisterRoutes(apiGroup, authMiddleware)
	matchHandler.RegisterRoutes(apiGroup, authMiddleware)
	evalHandler.RegisterRoutes(apiGroup, authMiddleware)
	courseHandler.RegisterRoutes(apiGroup, authMiddleware)

	// 6. Start HTTP Server
	addr := ":" + serverPort
	log.Printf("Servidor escoltant a http://localhost%s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Error fatal iniciant el servidor HTTP: %v", err)
	}
}
