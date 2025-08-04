package service

import (
	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/storage/postgres"
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenDuration  = time.Hour * 24
	refreshTokenDuration = time.Hour * 24 * 7
)

type UserService struct {
	Storage *postgres.Storage
}

func NewUserService(s *postgres.Storage) *UserService {
	return &UserService{Storage: s}
}

func (s *UserService) Register(ctx context.Context, username, email, password string) (string, string, error) {
	// Хешируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	// Создаём пользователя в БД
	u := model.User{
		UserName:     username,
		Email:        email,
		Password:     string(hash),
		RefreshToken: "",
	}
	if err := s.Storage.CreateUser(ctx, u); err != nil {
		return "", "", err
	}

	// После создания — получаем пользователя из БД (чтобы взять ID)
	createdUser, err := s.Storage.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}

	// Генерируем JWT токены
	access, refresh, err := generateTokens(createdUser.ID)
	if err != nil {
		return "", "", err
	}
	// Обновляем refresh_token в БД
	err = s.Storage.UpdateRefreshToken(ctx, createdUser.ID, refresh)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, string, error) {
	u, err := s.Storage.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return "", "", errors.New("invalid credentials")
	}
	access, refresh, err := generateTokens(u.ID)
	if err != nil {
		return "", "", err
	}
	err = s.Storage.UpdateRefreshToken(ctx, u.ID, refresh)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (s *UserService) Refresh(ctx context.Context, refreshToken string) (string, error) {
	user, err := s.Storage.GetUserByRefresh(ctx, refreshToken)
	if err != nil {
		return "", err
	}
	access, err := generateJWT(user.ID, accessTokenDuration)
	if err != nil {
		return "", err
	}
	return access, nil
}

func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.Storage.GetUserByID(ctx, id)
}

// --- JWT helper ---
func generateJWT(userID uuid.UUID, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func generateTokens(userID uuid.UUID) (string, string, error) {
	access, err := generateJWT(userID, accessTokenDuration)
	if err != nil {
		return "", "", err
	}
	refresh, err := generateJWT(userID, refreshTokenDuration)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func ParseToken(tokenStr string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}
	claims := token.Claims.(jwt.MapClaims)
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return uuid.Nil, errors.New("user_id not found or not a string")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, errors.New("invalid user_id format")
	}
	return userID, nil
}
