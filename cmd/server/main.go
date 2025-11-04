// @title TechUp API
// @version 1.0
// @description Backend API for the TechUp university application.
// @host localhost:8080
// @BasePath /
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"techup/config"
	"techup/internal/account"
	"techup/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "techup/docs" // generated swagger files
)

func main() {
	// Initialize logger
	logger.Init()

	// Load environment variables
	if err := config.LoadEnv(); err != nil {
		logger.Log.Fatal().Err(err).Msg("cannot load environment")
	}

	// Connect to PostgreSQL
	db, err := config.NewPostgresPool()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()
	logger.Log.Info().Msg("Database connected")

	// Initialize repository, service and handler
	repo := account.NewRepository(db)
	svc := account.NewService(repo)
	handler := account.NewHandler(svc)

	// Initialize Gin engine
	r := gin.Default()

	// Public routes
	r.POST("/register", handler.Register) // User registration
	r.POST("/login", handler.Login)       // User login

	// Protected routes with JWT authentication
	auth := r.Group("/")
	auth.Use(account.AuthMiddleware())
	{
		auth.GET("/profile", handler.Profile) // Get current user profile
		auth.POST("/account/change-password", handler.ChangePassword)
		auth.PUT("/account/update", handler.UpdateProfile)
		auth.POST("/account/set-role", handler.SetRole)
	}

	// Swagger UI route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler)) // Swagger UI

	// Start server
	port := config.GetPort()
	logger.Log.Info().Str("port", port).Msg("Server started")
	r.Run(":" + port) // listen and serve
}
