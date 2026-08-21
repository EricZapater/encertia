package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/encertia/backend/internal/auth"
	"github.com/encertia/backend/internal/db"
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

	// 4. Initialize Domain Modules
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

	// Middleware
	authMiddleware := shared.AuthMiddleware(authHandler.TokenValidatorAdapter())

	// 5. Setup Gin Router
	ginMode := getEnv("GIN_MODE", gin.ReleaseMode)
	gin.SetMode(ginMode)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery(), shared.CORSMiddleware())

	// Health Check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "encertia-backend",
		})
	})

	// Register routes directly at root (/auth/*, /users/*) as per OpenAPI contract
	rootGroup := router.Group("")
	authHandler.RegisterRoutes(rootGroup, authMiddleware)
	userHandler.RegisterRoutes(rootGroup, authMiddleware)

	// Also support /api prefix for proxy convenience (/api/auth/*, /api/users/*)
	apiGroup := router.Group("/api")
	authHandler.RegisterRoutes(apiGroup, authMiddleware)
	userHandler.RegisterRoutes(apiGroup, authMiddleware)

	// 6. Start HTTP Server
	addr := ":" + serverPort
	log.Printf("Servidor escoltant a http://localhost%s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Error fatal iniciant el servidor HTTP: %v", err)
	}
}
