package main

import (
	"book-foto-art-back/internal/handler"
	"book-foto-art-back/internal/service"
	"book-foto-art-back/internal/storage"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	db := storage.InitDB()
	userService := service.NewUserService(db)
	h := handler.NewHandler(userService)

	//r := gin.Default()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Настройка CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://bookfoto.art", "http://localhost:4200"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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
