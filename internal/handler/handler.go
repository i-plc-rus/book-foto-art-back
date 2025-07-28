package handler

import (
	"book-foto-art-back/internal/service"
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	userService       *service.UserService
	collectionService *service.CollectionService
	uploadService     *service.UploadService
}

func NewHandler(
	userService *service.UserService,
	collectionService *service.CollectionService,
	uploadService *service.UploadService,
) *Handler {
	return &Handler{
		userService:       userService,
		collectionService: collectionService,
		uploadService:     uploadService,
	}
}

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		userID, err := service.ParseToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		c.Set("user_id", userID.String())
		c.Next()
	}
}

func (h *Handler) Refresh(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	user, err := h.userService.Storage.GetUserByRefresh(c.Request.Context(), input.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
	access, refresh, err := service.GenerateTokens(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token error"})
		return
	}
	_ = h.userService.Storage.UpdateRefreshToken(c.Request.Context(), user.ID, refresh)
	c.JSON(http.StatusOK, gin.H{"access_token": access, "refresh_token": refresh})
}

func (h *Handler) Register(c *gin.Context) {
	var input struct {
		UserName string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	token, err := h.userService.Register(c.Request.Context(), input.UserName, input.Email, input.Password)
	if err != nil {
		log.Printf("failed to register user: %v", err)
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"token": token})
}

func (h *Handler) Login(c *gin.Context) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	token, err := h.userService.Login(c.Request.Context(), input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) GetProfile(c *gin.Context) {
	// Получаем user_id из контекста
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	user, err := h.userService.GetProfile(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": user.ID, "email": user.Email})
}

func (h *Handler) CreateCollection(c *gin.Context) {
	var input struct {
		Name string `json:"name"`
		Date string `json:"date"` // Можно потом заменить на time.Time с парсингом
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Получаем user_id из контекста
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Парсим дату (пример)
	date, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	col, err := h.collectionService.CreateCollection(c.Request.Context(), userID, input.Name, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create collection"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": col.ID})
}

func (h *Handler) GetCollection(c *gin.Context) {
	collectionIDStr := c.Param("id")
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		log.Printf("Invalid collection ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	collection, err := h.collectionService.GetCollectionByID(c.Request.Context(), collectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get collection"})
		}
		return
	}

	c.JSON(http.StatusOK, collection)
}

func (h *Handler) ListCollections(c *gin.Context) {
	// Получаем user_id из контекста
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collections, err := h.collectionService.GetCollections(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get collections"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"collections": collections})
}

func (h *Handler) UploadFiles(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	collectionIDStr := c.GetString("collection_id")
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		log.Printf("Invalid collection ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("Failed to get multipart form: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get multipart form"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		log.Printf("No files provided: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files provided"})
		return
	}

	// ✅ Вызов сервиса один раз со всеми файлами
	results, err := h.uploadService.UploadFiles(c.Request.Context(), userID, collectionID, files)
	if err != nil {
		log.Printf("Failed to upload files: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": results})
}
