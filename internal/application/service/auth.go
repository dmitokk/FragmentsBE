package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/dmitokk/FragmentsBE/internal/application/dto"
	"github.com/dmitokk/FragmentsBE/internal/domain/entity"
	"github.com/dmitokk/FragmentsBE/internal/domain/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type AuthService struct {
	userRepo      repository.UserRepository
	jwtSecret     string
	googleConfig  *oauth2.Config
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret, googleClientID, googleClientSecret, googleRedirectURL string) *AuthService {
	googleConfig := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  googleRedirectURL,
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}

	return &AuthService{
		userRepo:     userRepo,
		jwtSecret:    jwtSecret,
		googleConfig: googleConfig,
	}
}

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &entity.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		Token: token,
		User: struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}{
			ID:    user.ID.String(),
			Email: user.Email,
		},
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		Token: token,
		User: struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}{
			ID:    user.ID.String(),
			Email: user.Email,
		},
	}, nil
}

func (s *AuthService) GoogleAuthURL(state string) string {
	return s.googleConfig.AuthCodeURL(state)
}

func (s *AuthService) GoogleCallback(ctx context.Context, req *dto.GoogleAuthRequest) (*dto.AuthResponse, error) {
	token, err := s.googleConfig.Exchange(ctx, req.Code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	userInfo, err := s.googleConfig.Client(ctx, token).Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer userInfo.Body.Close()

	body, err := io.ReadAll(userInfo.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info: %w", err)
	}

	var googleUser struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}

	if err := json.Unmarshal(body, &googleUser); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	user, err := s.userRepo.GetByGoogleID(ctx, googleUser.ID)
	if err != nil {
		user = &entity.User{
			ID:       uuid.New(),
			Email:    googleUser.Email,
			GoogleID: googleUser.ID,
		}

		err = s.userRepo.Create(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	jwtToken, err := s.generateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.AuthResponse{
		Token: jwtToken,
		User: struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}{
			ID:    user.ID.String(),
			Email: user.Email,
		},
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr, ok := claims["user_id"].(string)
		if !ok {
			return uuid.Nil, fmt.Errorf("invalid token claims")
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return uuid.Nil, fmt.Errorf("invalid user id in token: %w", err)
		}

		return userID, nil
	}

	return uuid.Nil, fmt.Errorf("invalid token")
}

func (s *AuthService) generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}