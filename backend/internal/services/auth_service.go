package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/net1io/zenbali/internal/config"
	"github.com/net1io/zenbali/internal/models"
	"github.com/net1io/zenbali/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountDisabled    = errors.New("account is disabled")
	ErrEmailExists        = errors.New("email already registered")
)

type AuthService struct {
	repos  *repository.Repositories
	config config.JWTConfig
}

func NewAuthService(repos *repository.Repositories, config config.JWTConfig) *AuthService {
	return &AuthService{
		repos:  repos,
		config: config,
	}
}

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	UserType string    `json:"user_type"` // "creator" or "admin"
	jwt.RegisteredClaims
}

func (s *AuthService) RegisterCreator(ctx context.Context, req *models.CreatorRegisterRequest) (*models.Creator, error) {
	// Check if email exists
	existing, err := s.repos.Creator.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailExists
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	creator := &models.Creator{
		Name:             req.Name,
		OrganizationName: req.OrganizationName,
		Email:            req.Email,
		Mobile:           req.Mobile,
		PasswordHash:     string(hash),
		IsActive:         true,
	}

	if err := s.repos.Creator.Create(ctx, creator); err != nil {
		return nil, err
	}

	return creator, nil
}

func (s *AuthService) LoginCreator(ctx context.Context, req *models.CreatorLoginRequest) (*models.Creator, string, error) {
	creator, err := s.repos.Creator.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", err
	}
	if creator == nil {
		return nil, "", ErrInvalidCredentials
	}

	if !creator.IsActive {
		return nil, "", ErrAccountDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(creator.PasswordHash), []byte(req.Password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.generateToken(creator.ID, creator.Email, "creator")
	if err != nil {
		return nil, "", err
	}

	return creator, token, nil
}

func (s *AuthService) LoginAdmin(ctx context.Context, req *models.AdminLoginRequest) (*models.Admin, string, error) {
	admin, err := s.repos.Admin.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", err
	}
	if admin == nil {
		return nil, "", ErrInvalidCredentials
	}

	if !admin.IsActive {
		return nil, "", ErrAccountDisabled
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	token, err := s.generateToken(admin.ID, admin.Email, "admin")
	if err != nil {
		return nil, "", err
	}

	return admin, token, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *AuthService) GetCreatorByID(ctx context.Context, id uuid.UUID) (*models.Creator, error) {
	return s.repos.Creator.GetByID(ctx, id)
}

func (s *AuthService) GetAdminByID(ctx context.Context, id uuid.UUID) (*models.Admin, error) {
	return s.repos.Admin.GetByID(ctx, id)
}

func (s *AuthService) generateToken(userID uuid.UUID, email, userType string) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Email:    email,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.config.ExpiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "zenbali",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.Secret))
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (s *AuthService) EnsureDefaultAdmin(ctx context.Context, email, password string) error {
	hash, err := s.HashPassword(password)
	if err != nil {
		return err
	}
	return s.repos.Admin.EnsureDefaultAdmin(ctx, email, hash)
}
