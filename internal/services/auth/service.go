package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"auth-service/pkg/domain"
)

type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	SaveAuthCode(ctx context.Context, authCode *domain.AuthCode) error
	GetAuthCode(ctx context.Context, userID string) (*domain.AuthCode, error)
}

type KafkaProducer interface {
	SendEmail(ctx context.Context, email domain.EmailMessage) error
	SendUserRegistered(ctx context.Context, event domain.UserRegisteredEvent) error
}

type Service struct {
	userRepo      UserRepository
	kafkaProducer KafkaProducer
	jwtSecret     string
	emailFrom     string
}

func NewAuthService(userRepo UserRepository, kafkaProducer KafkaProducer, jwtSecret, emailFrom string) *Service {
	return &Service{
		userRepo:      userRepo,
		kafkaProducer: kafkaProducer,
		jwtSecret:     jwtSecret,
		emailFrom:     emailFrom,
	}
}

func (s *Service) Register(ctx context.Context, req domain.RegisterRequest) (*domain.User, error) {
	existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	if err := s.generateAndSendCode(ctx, user); err != nil {
		return nil, err
	}

	userEvent := domain.UserRegisteredEvent{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	}
	if err := s.kafkaProducer.SendUserRegistered(ctx, userEvent); err != nil {
		fmt.Printf("Failed to send user registered event: %v\n", err)
	}

	user.Password = ""
	return user, nil
}

func (s *Service) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := s.generateAndSendCode(ctx, user); err != nil {
		return nil, err
	}

	user.Password = ""
	return &domain.AuthResponse{
		User: *user,
	}, nil
}

func (s *Service) VerifyCode(ctx context.Context, req domain.VerifyCodeRequest) (*domain.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid email")
	}

	authCode, err := s.userRepo.GetAuthCode(ctx, user.ID)
	if err != nil {
		return nil, errors.New("invalid or expired code")
	}

	if authCode.ExpiresAt.Before(time.Now()) {
		return nil, errors.New("code expired")
	}

	if authCode.Code != req.Code {
		return nil, errors.New("invalid code")
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return nil, err
	}

	user.Password = ""
	return &domain.AuthResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *Service) generateAndSendCode(ctx context.Context, user *domain.User) error {
	code, err := s.generateCode(4)
	if err != nil {
		return err
	}

	authCode := &domain.AuthCode{
		UserID:    user.ID,
		Code:      code,
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	if err := s.userRepo.SaveAuthCode(ctx, authCode); err != nil {
		return err
	}

	emailMsg := domain.EmailMessage{
		ID:      uuid.New().String(),
		From:    s.emailFrom,
		To:      user.Email,
		Subject: "Your Authentication Code",
		Body:    fmt.Sprintf("Your authentication code is: %s. It will expire in 15 minutes.", code),
	}

	return s.kafkaProducer.SendEmail(ctx, emailMsg)
}

func (s *Service) generateCode(length int) (string, error) {
	const digits = "0123456789"
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}
	return string(code), nil
}

func (s *Service) generateJWT(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
