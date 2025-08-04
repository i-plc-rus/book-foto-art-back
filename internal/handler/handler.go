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

// Refresh godoc
// @Summary      Обновление токена
// @Description  Обновляет access и refresh токены
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input body model.RefreshRequest true "Refresh токен"
// @Success      200 {object} model.RefreshResponse
// @Failure      401 {object} model.ErrorMessage
// @Router       /auth/refresh [post]
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

// Register godoc
// @Summary      Регистрация пользователя
// @Description  Создаёт нового пользователя
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input body model.RegisterRequest true "Данные для регистрации"
// @Success      201 {object} model.TokenResponse
// @Failure      400 {object} model.ErrorMessage
// @Router       /auth/register [post]
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

// Login godoc
// @Summary      Аутентификация пользователя
// @Description  Аутентифицирует пользователя
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input body model.LoginRequest true "Данные для входа"
// @Success      200 {object} model.TokenResponse
// @Failure      401 {object} model.ErrorMessage
// @Router       /auth/login [post]
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

// GetProfile godoc
// @Summary      Получить профиль пользователя
// @Description  Возвращает данные профиля текущего пользователя
// @Tags         Profile
// @Accept       json
// @Produce      json
// @Success      200 {object} model.ProfileResponse
// @Failure      400 {object} model.ErrorMessage
// @Router       /profile/ [get]
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

// CreateCollection godoc
// @Summary      Создать коллекцию
// @Description  Создаёт новую коллекцию для пользователя
// @Tags         Collection
// @Accept       json
// @Produce      json
// @Param        input body model.CreateCollectionRequest true "Данные коллекции"
// @Success      201 {object} model.CreateCollectionResponse
// @Failure      400 {object} model.ErrorMessage
// @Router       /collection/create [post]
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
		log.Printf("Failed to create collection: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create collection"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": col.ID})
}

// GetCollectionInfo godoc
// @Summary      Получить информацию о коллекции
// @Description  Возвращает информацию о коллекции по ID
// @Tags         Collection
// @Accept       json
// @Produce      json
// @Param        id path string true "ID коллекции"
// @Success      200 {object} model.CollectionInfoResponse
// @Failure      404 {object} model.ErrorMessage
// @Router       /collection/{id} [get]
func (h *Handler) GetCollectionInfo(c *gin.Context) {
	// Получаем user_id из контекста
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получаем collection_id из URL
	collectionIDStr := c.Param("id")
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		log.Printf("Invalid collection ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	collection, err := h.collectionService.GetCollectionInfo(c.Request.Context(), userID, collectionID)
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

// GetCollectionPhotos godoc
// @Summary      Получить фотографии коллекции
// @Description  Возвращает список фотографий в коллекции пользователя с возможностью сортировки.
// @Tags         Collection
// @Accept       json
// @Produce      json
// @Param        id   path   string true  "ID коллекции"
// @Param        sort query  string false "Сортировка. Возможные значения: uploaded_new (по дате загрузки, новые сверху), uploaded_old (по дате загрузки, старые сверху), name_az (по имени файла, A-Z), name_za (по имени файла, Z-A), random (случайный порядок). По умолчанию: uploaded_new"
// @Success      200  {object} model.CollectionPhotosResponse
// @Failure      404  {object} model.ErrorMessage
// @Router       /collection/{id}/photos [get]
func (h *Handler) GetCollectionPhotos(c *gin.Context) {
	// Получаем user_id из контекста
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получаем параметры из URL
	collectionIDStr := c.Param("id")
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		log.Printf("Invalid collection ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}
	sortParam := c.Query("sort")

	photos, sort, err := h.collectionService.GetCollectionPhotos(c.Request.Context(), userID, collectionID, sortParam)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get collection photos"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": photos,
		"sort":  sort,
	})
}

// ListCollections godoc
// @Summary      Список коллекций
// @Description  Возвращает список коллекций пользователя
// @Tags         Collection
// @Accept       json
// @Produce      json
// @Success      200 {object} model.CollectionsListResponse
// @Failure      400 {object} model.ErrorMessage
// @Router       /collection/list [get]
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

// DeleteCollection godoc
// @Summary      Удалить коллекцию
// @Description  Удаление коллекции по ID
// @Tags         Collection
// @Accept       json
// @Produce      json
// @Param        id path string true "ID коллекции"
// @Success      200 {object} model.BooleanResponse
// @Failure      404 {object} model.ErrorMessage
// @Failure      400 {object} model.ErrorMessage
// @Router       /collection/{id} [delete]
func (h *Handler) DeleteCollection(c *gin.Context) {
	// Получаем user_id из контекста
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получаем collection_id из URL
	collectionIDStr := c.Param("id")
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		log.Printf("Invalid collection ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	err = h.collectionService.DeleteCollection(c.Request.Context(), userID, collectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete collection"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UploadFiles godoc
// @Summary      Загрузка файлов
// @Description  Загружает файлы в коллекцию
// @Tags         Upload
// @Accept       multipart/form-data
// @Produce      json
// @Param        collection_id formData string true "ID коллекции"
// @Param        files formData file true "Файлы для загрузки"
// @Success      200 {object} model.UploadFilesResponse
// @Failure      400 {object} model.ErrorMessage
// @Router       /upload/files [post]
func (h *Handler) UploadFiles(c *gin.Context) {
	// Получаем user_id из контекста
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получаем collection_id из формы
	collectionIDStr := c.PostForm("collection_id")
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

// UpdateCollectionCover godoc
// @Summary      Обновить обложку коллекции
// @Description  Изменяет обложку коллекции на фотографию из этой коллекции
// @Tags         Collection
// @Accept       json
// @Produce      json
// @Param        id path string true "ID коллекции"
// @Param        input body model.UpdateCollectionCoverRequest true "Данные для обновления обложки"
// @Success      200 {object} model.BooleanResponse
// @Failure      404 {object} model.ErrorMessage
// @Router       /collection/{id}/cover [put]
func (h *Handler) UpdateCollectionCover(c *gin.Context) {
	// Получаем user_id из контекста
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получаем collection_id из URL
	collectionIDStr := c.Param("id")
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		log.Printf("Invalid collection ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	var input struct {
		UploadedPhotoIDStr string `json:"photo_id"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	uploadedPhotoID, err := uuid.Parse(input.UploadedPhotoIDStr)
	if err != nil {
		log.Printf("Invalid photo ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid photo ID"})
		return
	}

	err = h.collectionService.UpdateCollectionCover(c.Request.Context(), userID, collectionID, uploadedPhotoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Photo not found in collection"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update collection cover"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
