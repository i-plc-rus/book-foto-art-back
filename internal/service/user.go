package service

import (
	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/storage"
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	Storage *storage.Storage
}

func NewUserService(s *storage.Storage) *UserService {
	return &UserService{Storage: s}
}

func (s *UserService) Register(ctx context.Context, email, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.Storage.CreateUser(ctx, model.User{Email: email, Password: string(hash)})
}

func (s *UserService) Login(ctx context.Context, email, password string) (string, error) {
	u, err := s.Storage.GetUserByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}
	return generateJWT(u.ID, time.Minute*15)
}

func (s *UserService) GetProfile(ctx context.Context, id int64) (*model.User, error) {
	return s.Storage.GetUserByID(ctx, id)
}

// --- JWT helper ---

func ParseToken(tokenStr string) (int64, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}
	claims := token.Claims.(jwt.MapClaims)
	return int64(claims["user_id"].(float64)), nil
}

func generateJWT(userID int64, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func GenerateTokens(userID int64) (access string, refresh string, err error) {
	access, err = generateJWT(userID, time.Minute*15)
	if err != nil {
		return
	}
	refresh, err = generateJWT(userID, time.Hour*24*7)
	return
}
