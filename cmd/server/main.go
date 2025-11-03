package main

import (
	"log"
	"techup/config"
	"techup/internal/account"

	"github.com/gin-gonic/gin"
)

func main() {
	// Загружаем переменные окружения
	config.LoadEnv()

	// Подключаемся к базе
	db, err := config.NewPostgresPool()
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer db.Close()

	// Инициализация репозитория, сервиса и хендлера
	repo := account.NewRepository(db)
	svc := account.NewService(repo)
	handler := account.NewHandler(svc)

	r := gin.Default()

	// Публичные маршруты
	r.POST("/register", handler.Register)
	r.POST("/login", handler.Login)

	// Защищённые маршруты
	auth := r.Group("/")
	auth.Use(account.AuthMiddleware()) // middleware проверяет JWT
	{
		auth.GET("/profile", handler.Profile)
	}

	port := config.GetPort()
	log.Printf("Server running on port %s", port)
	r.Run(":" + port)
}
