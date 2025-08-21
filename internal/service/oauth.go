package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yandex"

	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/storage/postgres"
)

type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewYandexOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("YANDEX_CLIENT_ID"),
		ClientSecret: os.Getenv("YANDEX_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("YANDEX_REDIRECT_URL"),
		Scopes:       []string{"login:info", "login:email", "login:avatar"},
		Endpoint:     yandex.Endpoint,
	}
}

type YandexOAuthService struct {
	oauthConfig *oauth2.Config
	userStorage *postgres.Storage
}

type YandexUserInfo struct {
	ID            string `json:"id"`
	Login         string `json:"login"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	DisplayName   string `json:"display_name"`
	RealName      string `json:"real_name"`
	Sex           string `json:"sex"`
	Email         string `json:"default_email"`
	IsAvatarEmpty bool   `json:"is_avatar_empty"`
	AvatarID      string `json:"default_avatar_id"`
}

func NewYandexOAuthService(oauthConfig *oauth2.Config, userStorage *postgres.Storage) *YandexOAuthService {
	return &YandexOAuthService{
		oauthConfig: oauthConfig,
		userStorage: userStorage,
	}
}

func (s *YandexOAuthService) GetAuthURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state)
}

func (s *YandexOAuthService) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	return s.oauthConfig.Exchange(ctx, code)
}

func (s *YandexOAuthService) GetUserInfo(ctx context.Context, token *oauth2.Token) (*YandexUserInfo, error) {
	client := s.oauthConfig.Client(ctx, token)

	resp, err := client.Get("https://login.yandex.ru/info")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info, status: %d", resp.StatusCode)
	}

	var userInfo YandexUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

func (s *YandexOAuthService) AuthenticateOrCreateUser(ctx context.Context, yandexUser *YandexUserInfo) (*model.User, string, string, error) {
	// Если пользователь существует, генерируем новые токены
	existingUser, err := s.userStorage.GetUserByEmail(ctx, yandexUser.Email)
	if err == nil && existingUser != nil {
		accessToken, refreshToken, err := GenerateTokens(existingUser.ID)
		if err != nil {
			return nil, "", "", fmt.Errorf("failed to generate tokens: %w", err)
		}
		return existingUser, accessToken, refreshToken, nil
	}

	// Создаем нового пользователя
	userID := uuid.New()
	username := yandexUser.Login
	if username == "" {
		username = yandexUser.Email
	}
	newUser := &model.User{
		ID:       userID,
		UserName: username,
		Email:    yandexUser.Email,
	}
	if err := s.userStorage.CreateUser(ctx, *newUser); err != nil {
		return nil, "", "", fmt.Errorf("failed to create user: %w", err)
	}

	// Генерируем токены для нового пользователя
	accessToken, refreshToken, err := GenerateTokens(userID)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	return newUser, accessToken, refreshToken, nil
}
