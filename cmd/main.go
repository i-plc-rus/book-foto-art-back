package main

import (
	"book-foto-art-back/internal/handler"
	"book-foto-art-back/internal/service"
	"book-foto-art-back/internal/storage"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	db := storage.InitDB()
	userService := service.NewUserService(db)
	h := handler.NewHandler(userService)

	r := gin.Default()

	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		//auth.GET("/yandex/login", h.YandexLogin)
		//auth.GET("/yandex/callback", h.YandexCallback)
	}

	profile := r.Group("/profile")
	{
		profile.Use(h.AuthMiddleware())
		profile.GET("/", h.GetProfile)
	}

	log.Fatal(r.Run(":8080"))
}
