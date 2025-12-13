package main

import (
	"techup/config"
	"techup/internal/account"
	"techup/internal/health"
	"techup/internal/logger"
	maps "techup/internal/map"
	"techup/internal/schedule"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "techup/docs"
)

func main() {
	logger.Init()

	if err := config.LoadEnv(); err != nil {
		logger.Log.Fatal().Err(err).Msg("cannot load environment")
	}

	db, err := config.NewPostgresPool()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()
	logger.Log.Info().Msg("database connected")

	// Repos & services
	accountRepo := account.NewRepository(db)
	accountSvc := account.NewService(accountRepo)
	accountHandler := account.NewHandler(accountSvc)

	scheduleRepo := schedule.NewRepository(db)
	scheduleSvc := schedule.NewService(scheduleRepo)
	scheduleHandler := schedule.NewHandler(scheduleSvc)

	mapsRepo := maps.NewRepository(db)
	mapsSvc := maps.NewService(mapsRepo)
	mapsHandler := maps.NewHandler(mapsSvc)

	// Gin router
	r := gin.New()

	// --- CORS ---
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			allowed := []string{
				"http://localhost:5173",
				"https://leanqha.github.io",
				"https://nonimpregnated-turner-acknowledgingly.ngrok-free.dev",
			}
			for _, o := range allowed {
				if o == origin {
					return true
				}
			}
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// OPTIONS preflight handler (для всех путей)
	r.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(200)
	})

	// Logger / Recovery
	r.Use(logger.RecoveryLogger())
	r.Use(logger.RequestIDMiddleware())
	r.Use(logger.GinLoggerMiddleware())

	// Health
	r.GET("/api/v1/health", health.Handler)

	api := r.Group("/api/v1")

	// Account routes
	accountGroup := api.Group("/account")
	accountGroup.POST("/register", accountHandler.Register)
	accountGroup.POST("/login", accountHandler.Login)
	accountGroup.POST("/refresh", accountHandler.Refresh)

	secureAccount := accountGroup.Group("/secure")
	secureAccount.Use(account.AuthMiddleware())
	{
		secureAccount.GET("/profile", accountHandler.Profile)
		secureAccount.POST("/change-password", accountHandler.ChangePassword)
		secureAccount.PUT("/update", accountHandler.UpdateProfile)
		secureAccount.POST("/logout", accountHandler.Logout)
	}

	// Schedule routes
	scheduleGroup := api.Group("/schedule")
	{
		scheduleGroup.GET("/lessons", scheduleHandler.ListLessons)
		scheduleGroup.GET("/groups", scheduleHandler.ListGroups)
		scheduleGroup.GET("/faculties", scheduleHandler.ListFaculties)
	}
	scheduleGroup.GET("/lessons/:id/note", scheduleHandler.GetLessonNote)
	scheduleGroup.POST("/lessons/:id/note", scheduleHandler.AddLessonNote)

	// Map routes
	mapGroup := api.Group("/map")
	{
		mapGroup.GET("/search", mapsHandler.SearchRooms)
		mapGroup.GET("/buildings", mapsHandler.GetBuildings)
		mapGroup.GET("/path/:start/:end", mapsHandler.GetPath)
	}

	// Admin routes
	adminGroup := api.Group("/admin")
	adminGroup.Use(account.AuthMiddleware(), account.RequireRole("admin"))
	{
		adminGroup.DELETE("/account/:id", accountHandler.DeleteAccount)
		adminGroup.POST("/set-role", accountHandler.SetRole)

		adminGroup.POST("/faculty", scheduleHandler.AddFaculty)
		adminGroup.PUT("/faculty/:id", scheduleHandler.UpdateFaculty)
		adminGroup.DELETE("/faculty/:id", scheduleHandler.DeleteFaculty)

		adminGroup.POST("/group", scheduleHandler.AddGroup)
		adminGroup.PUT("/group/:id", scheduleHandler.UpdateGroup)
		adminGroup.DELETE("/group/:id", scheduleHandler.DeleteGroup)

		adminGroup.POST("/room", mapsHandler.AddRoom)
		adminGroup.POST("/lesson", scheduleHandler.AddLesson)
		adminGroup.PUT("/lesson/:id", scheduleHandler.UpdateLesson)
		adminGroup.DELETE("/lesson/:id", scheduleHandler.DeleteLesson)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Run server
	port := config.GetPort()
	if err := r.Run(":" + port); err != nil {
		logger.Log.Fatal().Err(err).Msg("cannot start server")
	}
}
