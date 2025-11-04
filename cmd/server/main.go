package main

import (
	"log"
	"techup/config"
	"techup/internal/account"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()

	db, err := config.NewPostgresPool()
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer db.Close()

	repo := account.NewRepository(db)
	svc := account.NewService(repo)
	handler := account.NewHandler(svc)

	r := gin.Default()

	r.POST("/register", handler.Register)
	r.POST("/login", handler.Login)

	auth := r.Group("/")
	auth.Use(account.AuthMiddleware())
	{
		auth.GET("/profile", handler.Profile)
	}

	port := config.GetPort()
	log.Printf("Server running on port %s", port)
	r.Run(":" + port)
}
