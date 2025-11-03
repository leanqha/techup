package main

import (
	"techup/config"
	"techup/internal/account"

	"github.com/gin-gonic/gin"
)

func main() {
	db, _ := config.NewPostgresPool()
	repo := account.NewRepository(db)
	service := account.NewService(repo)
	handler := account.NewHandler(service)

	r := gin.Default()

	// Публичные маршруты
	r.POST("/register", handler.Register)
	r.POST("/login", handler.Login)

	// Защищённые маршруты
	auth := r.Group("/")
	auth.Use(account.AuthMiddleware())
	{
		auth.GET("/profile", handler.Profile)
	}

	r.Run(":8080")
}
