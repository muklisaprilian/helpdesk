package services

import (
	"errors"
	"fmt"

	"github.com/simplejourney/helpdesk/internal/middleware"
	"github.com/simplejourney/helpdesk/internal/models"
	"github.com/simplejourney/helpdesk/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users      *repository.UserRepository
	jwtSecret  string
	jwtExpiryH int
}

func NewAuthService(users *repository.UserRepository, jwtSecret string, jwtExpiryH int) *AuthService {
	return &AuthService{users: users, jwtSecret: jwtSecret, jwtExpiryH: jwtExpiryH}
}

func (s *AuthService) Register(req models.RegisterRequest) (*models.AuthResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user, err := s.users.Create(req.Name, req.Email, string(hash), models.RoleRequester)
	if err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return nil, fmt.Errorf("email already registered")
		}
		return nil, err
	}
	token, err := middleware.GenerateToken(s.jwtSecret, s.jwtExpiryH, *user)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return &models.AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) Login(req models.LoginRequest) (*models.AuthResponse, error) {
	user, err := s.users.FindByEmail(req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, fmt.Errorf("invalid email or password")
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}
	token, err := middleware.GenerateToken(s.jwtSecret, s.jwtExpiryH, *user)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return &models.AuthResponse{Token: token, User: *user}, nil
}

func (s *AuthService) Me(userID string) (*models.User, error) {
	user, err := s.users.FindByID(userID)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	return user, nil
}
