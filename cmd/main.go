package main

import (
	_ "book-foto-art-back/docs"
	"book-foto-art-back/internal/handler"
	"book-foto-art-back/internal/service"
	"book-foto-art-back/internal/storage/postgres"
	"book-foto-art-back/internal/storage/s3"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	// "github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title BookFotoArt API
// @version 1.0
// @description API для сервиса BookFotoArt
// @host api.bookfoto.art
// @BasePath /
// @schemes https
func main() {

	// // Загрузка переменных окружения (local)
	// if err := godotenv.Load(".env.local"); err != nil {
	// 	log.Println("Error loading .env.local file")
	// }

	// БД
	pgStorage := postgres.InitDB()

	// S3 Storage
	s3Storage, err := s3.NewS3Storage(s3.S3Config{
		Region:          os.Getenv("AWS_REGION"),
		Bucket:          os.Getenv("AWS_BUCKET"),
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Endpoint:        os.Getenv("AWS_ENDPOINT"),
	})
	if err != nil {
		log.Fatalf("Failed to initialize S3 storage: %v", err)
	}

	// Сервисы
	userService := service.NewUserService(pgStorage)
	oauthService := service.NewYandexOAuthService(service.NewYandexOAuthConfig(), pgStorage)
	collectionService := service.NewCollectionService(pgStorage, s3Storage)
	uploadService := service.NewUploadService(pgStorage, s3Storage)

	// Обработчик
	h := handler.NewHandler(
		userService,
		oauthService,
		collectionService,
		uploadService,
	)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Printf("panic recovered: %s\n", err)
		} else if err, ok := recovered.(error); ok {
			log.Printf("panic recovered: %v\n", err)
		} else {
			log.Printf("panic recovered: unknown error: %v\n", recovered)
		}
		c.AbortWithStatusJSON(500, gin.H{"error": "internal server error"})
	}))

	// Настройка CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://bookfoto.art", "http://localhost:4200"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Авторизация
	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.GET("/yandex/login", h.YandexLogin)
		auth.GET("/yandex/callback", h.YandexCallback)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.POST("/reset-password", h.ResetPassword)
	}

	// Профиль
	profile := r.Group("/profile")
	{
		profile.Use(h.AuthMiddleware())
		profile.GET("/", h.GetProfile)
	}

	// Коллекции
	collection := r.Group("/collection")
	{
		collection.Use(h.AuthMiddleware())
		collection.POST("/create", h.CreateCollection)
		collection.GET("/list", h.ListCollections)
		collection.GET("/:id", h.GetCollectionInfo)
		collection.DELETE("/:id", h.DeleteCollection)
		collection.GET("/:id/photos", h.GetCollectionPhotos)
		collection.PUT("/:id/cover", h.UpdateCollectionCover)
		collection.DELETE("/photo/:photo_id", h.DeletePhoto)
		collection.PUT("/:id/publish", h.PublishCollection)
		collection.PUT("/:id/unpublish", h.UnpublishCollection)
		collection.GET("/short_link_info/:token", h.GetShortLinkInfo)
	}
	shortLink := r.Group("/s")
	{
		shortLink.GET("/:token", h.RedirectToPublicCollection)
	}
	public := r.Group("/public")
	{
		public.GET("collection/:token/photos", h.GetPublicCollection)
	}

	// Загрузка файлов
	upload := r.Group("/upload")
	{
		upload.Use(h.AuthMiddleware())
		upload.POST("/files", h.UploadFiles)
	}

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	log.Fatal(r.Run(":8080"))
}
