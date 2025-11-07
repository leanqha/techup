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
	"techup/internal/health"
	"techup/internal/logger"
	maps "techup/internal/map"
	"techup/internal/schedule"

	"github.com/gin-contrib/cors"
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
	logger.Log.Info().Msg("database connected")

	// Initialize account repository, service and handler
	accountRepo := account.NewRepository(db)
	accountSvc := account.NewService(accountRepo)
	accountHandler := account.NewHandler(accountSvc)

	// Initialize schedule repository, service and handler
	scheduleRepo := schedule.NewRepository(db)
	scheduleSvc := schedule.NewService(scheduleRepo)
	scheduleHandler := schedule.NewHandler(scheduleSvc)

	mapsRepo := maps.NewRepository(db)
	mapsSvc := maps.NewService(mapsRepo)
	mapsHandler := maps.NewHandler(mapsSvc)

	// Initialize Gin engine
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(logger.GinLoggerMiddleware())

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // фронтенд на Python или Vite
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/api/v1/health", health.Handler)

	// Базовая группа API
	api := r.Group("/api/v1")

	// ---------- Account routes ----------
	accountGroup := api.Group("/account")

	// --- Публичные ---
	accountGroup.POST("/register", accountHandler.Register)
	accountGroup.POST("/login", accountHandler.Login)
	accountGroup.POST("/refresh", accountHandler.Refresh)

	// --- Защищённые ---
	secureAccount := accountGroup.Group("/secure")
	secureAccount.Use(account.AuthMiddleware())
	{
		secureAccount.GET("/profile", accountHandler.Profile)
		secureAccount.POST("/change-password", accountHandler.ChangePassword)
		secureAccount.PUT("/update", accountHandler.UpdateProfile)
		secureAccount.POST("/set-role", accountHandler.SetRole)
	}

	// ---------- Schedule routes ----------
	scheduleGroup := api.Group("/schedule")
	scheduleGroup.Use(account.AuthMiddleware())
	{
		scheduleGroup.GET("/search", scheduleHandler.SearchSchedule)
	}

	// ---------- Map / Building routes ----------
	mapsGroup := api.Group("/map")
	mapsGroup.Use()
	{
		mapsGroup.GET("/buildings", mapsHandler.GetBuildings)
		mapsGroup.GET("/buildings/:building_id/rooms", mapsHandler.GetRoomsByBuilding)
		mapsGroup.GET("/shortest-path/:start/:end", mapsHandler.GetShortestPath)
	}

	// ---------- Admin routes ----------
	adminGroup := api.Group("/admin")
	adminGroup.Use(account.AuthMiddleware(), account.RequireRole("admin"))
	{
		adminGroup.POST("/faculty", scheduleHandler.AddFaculty)
		adminGroup.POST("/group", scheduleHandler.AddGroup)
		adminGroup.POST("/lesson", scheduleHandler.AddLesson)
	}

	// ---------- Swagger ----------
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	port := config.GetPort()
	err = r.Run(":" + port)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("cannot start server")
	}
}
