package service

import (
	"book-foto-art-back/internal/model"
	"book-foto-art-back/internal/storage/postgres"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/rvinnie/yookassa-sdk-go/yookassa"
	yoocommon "github.com/rvinnie/yookassa-sdk-go/yookassa/common"
	yoopayment "github.com/rvinnie/yookassa-sdk-go/yookassa/payment"
)

type PaymentService struct {
	paymentHandler *yookassa.PaymentHandler
	storage        *postgres.Storage
}

func NewPaymentService(storage *postgres.Storage) *PaymentService {
	return &PaymentService{
		paymentHandler: yookassa.NewPaymentHandler(yookassa.NewClient(
			os.Getenv("YOOKASSA_SHOP_ID"),
			os.Getenv("YOOKASSA_SECRET_KEY"),
		)),
		storage: storage,
	}
}

// const (
// 	monthPlan = "month"
// 	yearPlan  = "year"
// )

// type Plan string

var subscriptionPlans = map[string]float64{
	"month": 299.00,
	"year":  1943.00,
}

func (s *PaymentService) CreatePayment(ctx context.Context, userID uuid.UUID, plan string) (string, error) {

	// Получаем информацию о пользователе для заполнения чека по 54-ФЗ
	user, err := s.storage.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}

	paymentID := uuid.New()
	// Создаем платеж в статусе pending
	yooCassaPayment, err := s.paymentHandler.CreatePayment(&yoopayment.Payment{
		Amount: &yoocommon.Amount{
			Value:    strconv.FormatFloat(subscriptionPlans[plan], 'f', 2, 64),
			Currency: "RUB",
		},
		Confirmation: yoopayment.Redirect{
			Type:      "redirect",
			ReturnURL: fmt.Sprintf("%s/subscription/result?payment_id=%s", os.Getenv("FRONTEND_URL"), paymentID.String()),
		},
		Capture:       true,
		Description:   fmt.Sprintf("Подписка BookFotoArt для пользователя %s", user.UserName),
		PaymentMethod: yoopayment.PaymentMethodType("bank_card"),
		Receipt: &yoopayment.Receipt{
			Customer: &yoocommon.Customer{
				Email: user.Email,
			},
			Items: []*yoocommon.Item{
				{
					Description: fmt.Sprintf("Подписка BookFotoArt для пользователя %s", user.UserName),
					Amount: &yoocommon.Amount{
						Value:    strconv.FormatFloat(subscriptionPlans[plan], 'f', 2, 64),
						Currency: "RUB",
					},
					VatCode:        4,
					Quantity:       "1.00",
					PaymentSubject: "service",
					PaymentMode:    "full_prepayment",
				},
			},
		},
	})
	if err != nil {
		return "", err
	}

	// Сохраняем платеж в БД
	payment := &model.Payment{
		ID:                paymentID,
		UserID:            userID,
		YooKassaPaymentID: yooCassaPayment.ID,
		Plan:              plan,
		Amount:            subscriptionPlans[plan],
		Status:            string(yooCassaPayment.Status),
		CreatedAt:         *yooCassaPayment.CreatedAt,
	}
	if err := s.storage.CreatePayment(ctx, payment); err != nil {
		return "", err
	}

	// Получаем URL для оформления платежа
	if confirmationMap, ok := yooCassaPayment.Confirmation.(map[string]interface{}); ok {
		if confirmationURL, ok := confirmationMap["confirmation_url"].(string); ok {
			return confirmationURL, nil
		}
	}
	return "", fmt.Errorf("redirect URL not found")
}

func (s *PaymentService) GetPayment(ctx context.Context, userID uuid.UUID, paymentID string) (*model.Payment, error) {
	return s.storage.GetPayment(ctx, userID, paymentID)
}

func (s *PaymentService) ProcessWebhook(ctx context.Context, event map[string]interface{}) error {
	// Извлекаем объект платежа из события
	paymentData, ok := event["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid payment object")
	}

	// Извлекаем параметры из объекта платежа
	yookassaPaymentID, ok := paymentData["id"].(string)
	if !ok {
		return fmt.Errorf("invalid payment ID")
	}
	status, ok := paymentData["status"].(string)
	if !ok {
		return fmt.Errorf("invalid payment status")
	}

	switch status {
	case "pending":
		log.Printf("payment's %s status is %s\n", yookassaPaymentID, status)
		return nil
	case "succeeded":
		log.Printf("payment's %s status is %s\n", yookassaPaymentID, status)
		err := s.storage.ExtendUserSubscription(ctx, yookassaPaymentID, true)
		if err != nil {
			return err
		}
		err = s.storage.UpdatePaymentStatus(ctx, yookassaPaymentID, status)
		if err != nil {
			return err
		}
		return nil
	case "canceled":
		log.Printf("payment's %s status is %s\n", yookassaPaymentID, status)
		err := s.storage.UpdatePaymentStatus(ctx, yookassaPaymentID, status)
		if err != nil {
			return err
		}
		return nil
	default:
		return fmt.Errorf("invalid payment status")
	}
}
