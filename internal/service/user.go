package service

import (
	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/storage/postgres"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	accessTokenDuration  = time.Hour * 24
	refreshTokenDuration = time.Hour * 24 * 7
	resetTokenDuration   = time.Hour
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

func (s *UserService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.Storage.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user.ID != uuid.Nil {
		resetToken, err := generateJWT(user.ID, time.Hour)
		if err != nil {
			return err
		}
		err = s.Storage.UpdateResetToken(ctx, user.ID, resetToken)
		if err != nil {
			return err
		}
		err = s.sendPasswordResetEmail(email, resetToken)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *UserService) ResetPassword(ctx context.Context, resetToken, newPassword string) error {
	// Проверяем валидность токена
	userID, err := ParseToken(resetToken)
	if err != nil {
		return errors.New("invalid token")
	}
	user, err := s.Storage.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if user.ResetToken != resetToken {
		return errors.New("invalid token")
	}
	// Устанавливаем новый пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	err = s.Storage.ResetPassword(ctx, userID, string(hash))
	if err != nil {
		return err
	}
	// Сбрасываем ResetToken
	err = s.Storage.UpdateResetToken(ctx, userID, "")
	if err != nil {
		return err
	}
	return nil
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

// --- Password reset ---
func (s *UserService) sendPasswordResetEmail(email, resetToken string) error {
	// Создаём ссылку для сброса пароля
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", os.Getenv("FRONTEND_URL"), resetToken)

	// Формируем тему письма в base64
	encodedSubject := base64.StdEncoding.EncodeToString([]byte("Сброс пароля - BookFotoArt"))

	// Формируем тело письма
	body := fmt.Sprintf(`
	Вы запросили сброс пароля для вашего аккаунта в BookFotoArt.

	Для сброса пароля перейдите по ссылке, ссылка действительна в течение %.0f часа(ов):
	%s

	Если вы не запрашивали сброс пароля, проигнорируйте это письмо.

	С уважением,
	Команда BookFotoArt
	Сайт: https://bookfoto.art
	Email: info@bookfoto.art`, resetTokenDuration.Hours(), resetLink)

	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: =?UTF-8?B?%s?=\r\n"+
		"Reply-To: info@bookfoto.art\r\n"+
		"Return-Path: %s\r\n"+
		"Message-ID: <%s@bookfoto.art>\r\n"+
		"Date: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/plain; charset=UTF-8\r\n"+
		"Content-Transfer-Encoding: 8bit\r\n"+
		"X-Mailer: BookFotoArt/1.0\r\n"+
		"X-Priority: 3\r\n"+
		"X-MSMail-Priority: Normal\r\n"+
		"\r\n"+
		"%s\r\n", os.Getenv("SMTP_FROM"), email, encodedSubject, os.Getenv("SMTP_FROM"),
		time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700"),
		time.Now().Format("20060102150405"), body)

	// Формируем адрес сервера
	addr := fmt.Sprintf("%s:%s", os.Getenv("SMTP_HOST"), os.Getenv("SMTP_PORT"))

	// Настраиваем аутентификацию
	auth := smtp.PlainAuth("", os.Getenv("SMTP_USERNAME"), os.Getenv("SMTP_PASSWORD"), os.Getenv("SMTP_HOST"))

	// Отправляем письмо
	log.Printf("✅ Try to send reset password email\n")
	log.Printf("addr: %s\n", addr)
	log.Printf("from: %s\n", os.Getenv("SMTP_FROM"))
	log.Printf("email: %s\n", email)
	log.Printf("message length: %d bytes", len(message))
	err := smtp.SendMail(addr, auth, os.Getenv("SMTP_FROM"), []string{email}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	log.Printf("✅ Email sent successfully!")
	return nil
}
