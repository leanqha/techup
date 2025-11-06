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

	// Initialize Gin engine
	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // фронтенд на Python или Vite
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

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
		scheduleGroup.GET("/:group_name", scheduleHandler.GetScheduleByGroup)
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
