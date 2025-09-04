package handler

import (
	"book-foto-art-back/internal/service"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Handler struct {
	userService       *service.UserService
	paymentService    *service.PaymentService
	oauthService      *service.YandexOAuthService
	collectionService *service.CollectionService
	uploadService     *service.UploadService
}

func NewHandler(
	userService *service.UserService,
	paymentService *service.PaymentService,
	oauthService *service.YandexOAuthService,
	collectionService *service.CollectionService,
	uploadService *service.UploadService,
) *Handler {
	return &Handler{
		userService:       userService,
		paymentService:    paymentService,
		oauthService:      oauthService,
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

// func (h *Handler) SubscriptionMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		userIDStr := c.GetString("user_id")
// 		userID, err := uuid.Parse(userIDStr)
// 		if err != nil {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
// 			c.Abort()
// 			return
// 		}
// 		active, err := h.subscriptionService.CheckSubscription(c.Request.Context(), userID)
// 		if err != nil {
// 			log.Printf("Failed to check subscription: %v", err)
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check subscription"})
// 			c.Abort()
// 			return
// 		}
// 		if !active {
// 			c.JSON(http.StatusPaymentRequired, gin.H{
// 				"error":       "Subscription required",
// 				"message":     "Для использования сервиса необходима активная подписка",
// 				"payment_url": os.Getenv("FRONTEND_URL") + "/subscription",
// 			})
// 			c.Abort()
// 			return
// 		}
// 		c.Next()
// 	}
// }

// Register godoc
// @Summary      Регистрация нового пользователя
// @Description  Создаёт нового пользователя в системе. Проверяет уникальность email и username. При успешной регистрации возвращает access и refresh токены для аутентификации.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input body model.RegisterRequest true "Данные для регистрации пользователя"
// @Success      201 {object} model.TokenResponse "Пользователь успешно зарегистрирован"
// @Failure      400 {object} model.ErrorMessage "Неверный формат данных"
// @Failure      409 {object} model.ErrorMessage "Пользователь с таким email или username уже существует"
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

	// Регистрируем пользователя
	access, refresh, err := h.userService.Register(c.Request.Context(), input.UserName, input.Email, input.Password)
	if err != nil {
		log.Printf("failed to register user: %v", err)
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"access_token": access, "refresh_token": refresh})
}

// Login godoc
// @Summary      Аутентификация пользователя
// @Description  Аутентифицирует пользователя по email и паролю. При успешной аутентификации возвращает access и refresh токены для дальнейшего использования API.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input body model.LoginRequest true "Данные для входа в систему"
// @Success      200 {object} model.TokenResponse "Успешная аутентификация"
// @Failure      400 {object} model.ErrorMessage "Неверный формат данных"
// @Failure      401 {object} model.ErrorMessage "Неверные учетные данные"
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
	access, refresh, err := h.userService.Login(c.Request.Context(), input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": access, "refresh_token": refresh})
}

// YandexLogin godoc
// @Summary      Перенаправление на Яндекс OAuth
// @Description  Перенаправляет пользователя на страницу авторизации Яндекса
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Success      200 {object} model.YandexLoginResponse "Перенаправление на Яндекс"
// @Router       /auth/yandex/login [get]
func (h *Handler) YandexLogin(c *gin.Context) {
	// Создаем JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(10 * time.Minute).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}
	authURL := h.oauthService.GetAuthURL(tokenStr)
	c.JSON(http.StatusOK, gin.H{"url": authURL})
}

// YandexCallback godoc
// @Summary      Обработка callback от Яндекс OAuth
// @Description  Обрабатывает ответ от Яндекса после успешной авторизации
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        code query string true "Код авторизации от Яндекса"
// @Param        state query string true "State параметр для безопасности"
// @Success      200 {object} model.YandexCallbackResponse "Успешная аутентификация"
// @Failure      400 {object} model.ErrorMessage "Неверные параметры"
// @Failure      401 {object} model.ErrorMessage "Ошибка аутентификации"
// @Router       /auth/yandex/callback [get]
func (h *Handler) YandexCallback(c *gin.Context) {
	code := c.Query("code")
	stateToken := c.Query("state")
	if code == "" || stateToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing parameters"})
		return
	}

	// Проверяем JWT подпись и TTL
	_, err := jwt.Parse(stateToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		log.Printf("Failed to parse JWT token: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state token"})
		return
	}

	// Обмениваем код на токен
	oauthToken, err := h.oauthService.ExchangeCodeForToken(c.Request.Context(), code)
	if err != nil {
		log.Printf("Failed to exchange code for token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to authenticate with Yandex"})
		return
	}

	// Получаем информацию о пользователе
	yandexUser, err := h.oauthService.GetUserInfo(c.Request.Context(), oauthToken)
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info from Yandex"})
		return
	}

	// Аутентифицируем или создаем пользователя
	user, accessToken, refreshToken, err := h.oauthService.AuthenticateOrCreateUser(c.Request.Context(), yandexUser)
	if err != nil {
		log.Printf("Failed to authenticate/create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user authentication"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":       user.ID,
			"username": user.UserName,
			"email":    user.Email,
		},
	})
}

// Refresh godoc
// @Summary      Обновление токена доступа
// @Description  Обновляет access токен используя refresh токен. Refresh токен должен быть действительным и не истекшим. Возвращает новый access токен.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input body model.RefreshRequest true "Refresh токен для обновления"
// @Success      200 {object} model.RefreshResponse "Токен успешно обновлен"
// @Failure      400 {object} model.ErrorMessage "Неверный формат данных"
// @Failure      401 {object} model.ErrorMessage "Недействительный или истекший refresh токен"
// @Failure      500 {object} model.ErrorMessage "Внутренняя ошибка сервера"
// @Router       /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	access, err := h.userService.Refresh(c.Request.Context(), input.RefreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token error"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"access_token": access})
}

// ForgotPassword godoc
// @Summary      Запрос на сброс пароля
// @Description  Отправляет письмо для сброса пароля на указанный email.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input body model.ForgotPasswordRequest true "Email пользователя"
// @Success      200 {object} model.BooleanResponse "Письмо успешно отправлено"
// @Failure      400 {object} model.ErrorMessage "Неверный формат данных"
// @Failure      500 {object} model.ErrorMessage "Ошибка при отправке письма"
// @Router       /auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := h.userService.ForgotPassword(c.Request.Context(), input.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset password email"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ResetPassword godoc
// @Summary      Сброс пароля
// @Description  Сбрасывает пароль пользователя по одноразовой ссылке.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input body model.ResetPasswordRequest true "Новый пароль"
// @Param        token query string true "Одноразовый токен для сброса пароля"
// @Success      200 {object} model.BooleanResponse "Пароль успешно сброшен"
// @Failure      400 {object} model.ErrorMessage "Неверный формат данных"
// @Failure      500 {object} model.ErrorMessage "Ошибка при сбросе пароля"
// @Router       /auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	// Получаем resetToken из URL
	resetToken := c.Query("token")

	// Получаем новый пароль из тела запроса
	var input struct {
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Сбрасываем пароль
	err := h.userService.ResetPassword(c.Request.Context(), resetToken, input.NewPassword)
	if err != nil {
		if err.Error() == "invalid token" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reset token"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
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
	user, err := h.userService.GetUserByID(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": user.ID, "email": user.Email})
}

func (h *Handler) CreatePayment(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	plan := c.Request.FormValue("plan")
	if plan == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan"})
		return
	}
	confirmationURL, err := h.paymentService.CreatePayment(c.Request.Context(), userID, plan)
	if err != nil {
		log.Printf("Failed to create payment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"confirmation_url": confirmationURL})
}

func (h *Handler) GetPaymentStatus(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	paymentID := c.Param("payment_id")
	payment, err := h.paymentService.GetPayment(c.Request.Context(), userID, paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get payment"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": payment.Status})
}

func (h *Handler) YoomoneyWebhook(c *gin.Context) {
	// Читаем тело запроса
	body, err := c.GetRawData()
	if err != nil {
		log.Printf("Failed to read request body\n")
		return
	}

	// Проверяем подпись webhook'а
	signature := c.GetHeader("X-Content-HMAC")
	if !h.paymentService.VerifyWebhook(body, signature) {
		log.Printf("Invalid webhook signature\n")
		return
	}

	// Парсим событие
	var event map[string]interface{}
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Failed to parse event\n")
		return
	}

	// Обрабатываем событие
	if err := h.paymentService.ProcessWebhook(c.Request.Context(), event); err != nil {
		log.Printf("Failed to process webhook: %v\n", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": "true"})
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
		log.Printf("Failed to get collection info: %v\n", err)
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

// PublishCollection godoc
// @Summary      Публикация коллекции
// @Description  Публикует коллекцию по ID
// @Tags         Collection
// @Accept       json
// @Produce      json
// @Param        id path string true "ID коллекции"
// @Success      200 {object} model.PublishCollectionResponse
// @Failure      404 {object} model.ErrorMessage
// @Failure      400 {object} model.ErrorMessage
// @Router       /collection/{id}/publish [put]
func (h *Handler) PublishCollection(c *gin.Context) {
	// Получаем параметры из контекста и URL
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	collectionIDStr := c.Param("id")
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		log.Printf("Invalid collection ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	// Получаем публичную ссылку на коллекцию
	link, err := h.collectionService.PublishCollection(c.Request.Context(), userID, collectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		} else {
			log.Printf("Failed to publish collection: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish collection"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"link": link})
}

// RedirectToPublicCollection godoc
// @Summary      Перенаправление на публичную коллекцию
// @Description  Перенаправляет на публичную коллекцию по токену
// @Tags         Public
// @Accept       json
// @Produce      json
// @Param        token path string true "Токен короткой ссылки"
// @Success      302
// @Failure      404 {object} model.ErrorMessage
// @Router       /s/{token} [get]
func (h *Handler) RedirectToPublicCollection(c *gin.Context) {
	token := c.Param("token")
	link, err := h.collectionService.GetPublicCollectionLink(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "The link is not valid"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get public collection"})
		}
		return
	}
	c.Redirect(http.StatusFound, link)
}

// GetPublicCollection godoc
// @Summary      Получить публичную коллекцию
// @Description  Возвращает публичную коллекцию по токену с возможностью сортировки.
// @Tags         Collection
// @Accept       json
// @Produce      json
// @Param        token   path   string true  "Токен короткой ссылки"
// @Param        sort query  string false "Сортировка. Возможные значения: uploaded_new (по дате загрузки, новые сверху), uploaded_old (по дате загрузки, старые сверху), name_az (по имени файла, A-Z), name_za (по имени файла, Z-A), random (случайный порядок). По умолчанию: uploaded_new"
// @Success      200  {object} model.PublicCollectionResponse
// @Failure      404  {object} model.ErrorMessage
// @Router       /public/collection/{token}/photos [get]
func (h *Handler) GetPublicCollection(c *gin.Context) {
	// Получаем параметры из контекста и URL
	token := c.Param("token")
	sortParam := c.Query("sort")

	collection, photos, sort, err := h.collectionService.GetPublicCollection(
		c.Request.Context(), token, sortParam)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get collection photos"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username":            collection.UserName,
		"collection_name":     collection.Name,
		"date":                collection.Date,
		"cover_url":           collection.CoverURL,
		"cover_thumbnail_url": collection.CoverThumbnailURL,
		"count_photos":        collection.CountPhotos,
		"files":               photos,
		"sort":                sort,
	})
}

// UnpublishCollection godoc
// @Summary      Снятие публикации коллекции
// @Description  Снимает публикацию коллекции по ID
// @Tags         Collection
// @Accept       json
// @Produce      json
// @Param        id path string true "ID коллекции"
// @Success      200 {object} model.BooleanResponse
// @Failure      404 {object} model.ErrorMessage
// @Failure      400 {object} model.ErrorMessage
// @Router       /collection/{id}/unpublish [put]
func (h *Handler) UnpublishCollection(c *gin.Context) {
	// Получаем параметры из контекста и URL
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	collectionIDStr := c.Param("id")
	collectionID, err := uuid.Parse(collectionIDStr)
	if err != nil {
		log.Printf("Invalid collection ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection ID"})
		return
	}

	err = h.collectionService.UnpublishCollection(c.Request.Context(), userID, collectionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Collection is not published or does not exist"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unpublish collection"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetShortLinkInfo godoc
// @Summary      Получить информацию о короткой ссылке
// @Description  Возвращает информацию о короткой ссылке по токену
// @Tags         Collection
// @Accept       json
// @Produce      json
// @Param        token path string true "Токен короткой ссылки"
// @Success      200 {object} object{short_link_info=model.ShortLinkInfoResponse}
// @Failure      404 {object} model.ErrorMessage
// @Failure      400 {object} model.ErrorMessage
// @Router       /collection/short_link_info/{token} [get]
func (h *Handler) GetShortLinkInfo(c *gin.Context) {
	token := c.Param("token")
	shortLink, err := h.collectionService.GetShortLinkInfo(c.Request.Context(), token)
	if err != nil {
		log.Printf("Failed to get short link info: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get short link info"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"short_link_info": shortLink})
}

// DeletePhoto godoc
// @Summary      Удалить фотографию
// @Description  Удаляет фотографию из коллекции по ID
// @Tags         Collection
// @Accept       json
// @Produce      json
// @Param        photo_id path string true "ID фотографии"
// @Success      200 {object} model.BooleanResponse
// @Failure      404 {object} model.ErrorMessage
// @Failure      400 {object} model.ErrorMessage
// @Router       /collection/photo/{photo_id} [delete]
func (h *Handler) DeletePhoto(c *gin.Context) {
	// Получаем user_id из контекста
	userIDStr := c.GetString("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получаем photo_id из URL
	photoIDStr := c.Param("photo_id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		log.Printf("Invalid photo ID: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid photo ID"})
		return
	}

	// Удаляем фотографию из коллекции
	err = h.collectionService.DeletePhoto(c.Request.Context(), userID, photoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Photo not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete photo"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ListCollections godoc
// @Summary      Список коллекций
// @Description  Возвращает список коллекций пользователя
// @Tags         Collection
// @Accept       json
// @Produce      json
// @Param        search query string false "Поиск по названию коллекции"
// @Success      200 {object} model.CollectionsListResponse
// @Failure      400 {object} model.ErrorMessage
// @Failure      500 {object} model.ErrorMessage
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
	searchParam := c.Query("search")

	collections, err := h.collectionService.GetCollections(c.Request.Context(), userID, searchParam)
	if err != nil {
		log.Printf("Failed to get collections: %v\n", err)
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
