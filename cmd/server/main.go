// @title TechUp API
// @version 1.0
// @description Backend API for the TechUp university application.
// @host localhost:8080
// @BasePath /api/v1
// @schemes http

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

	accountRepo := account.NewRepository(db)
	accountSvc := account.NewService(accountRepo)
	accountHandler := account.NewHandler(accountSvc)

	scheduleRepo := schedule.NewRepository(db)
	scheduleSvc := schedule.NewService(scheduleRepo)
	scheduleHandler := schedule.NewHandler(scheduleSvc)

	mapsRepo := maps.NewRepository(db)
	mapsSvc := maps.NewService(mapsRepo)
	mapsHandler := maps.NewHandler(mapsSvc)

	r := gin.New()

	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://mytechup.ru"
		},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"},
	}))

	r.Use(func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Status(204)
			c.Abort()
			return
		}
		c.Next()
	})

	r.Use(logger.RecoveryLogger())
	r.Use(logger.RequestIDMiddleware())
	r.Use(logger.GinLoggerMiddleware())

	r.GET("/api/v1/health", health.Handler)

	api := r.Group("/api/v1")

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

	scheduleGroup := api.Group("/schedule")
	{
		scheduleGroup.GET("/lessons", scheduleHandler.GetLessons)
		scheduleGroup.GET("/groups", scheduleHandler.ListGroups)
		scheduleGroup.GET("/faculties", scheduleHandler.ListFaculties)
		scheduleGroup.GET("/lessons/:id/note", scheduleHandler.GetLessonNote)
		scheduleGroup.POST("/lessons/:id/note", scheduleHandler.AddLessonNote)
		scheduleGroup.GET("/lessons/search", scheduleHandler.SearchLessons)
		scheduleGroup.GET("/teachers", scheduleHandler.GetTeachers)
		scheduleGroup.GET("/classrooms", scheduleHandler.GetClassrooms)
	}

	mapGroup := api.Group("/map")
	{
		mapGroup.GET("/search", mapsHandler.SearchRooms)
		mapGroup.GET("/buildings", mapsHandler.GetBuildings)
		mapGroup.GET("/path/:start/:end", mapsHandler.GetPath)
	}

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

		adminGroup.POST("/schedule/import", scheduleHandler.ImportSchedule)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	port := config.GetPort()
	err = r.Run(":" + port)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("cannot start server")
	}
}
