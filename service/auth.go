package service

import (
	"context"
	"time"

	"github.com/Rakesh2908/taskflow/config"
	"github.com/Rakesh2908/taskflow/dto"
	"github.com/Rakesh2908/taskflow/models"
	repo "github.com/Rakesh2908/taskflow/repository"
	"github.com/Rakesh2908/taskflow/utils"
	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
}

type authService struct {
	cfg   config.Config
	users repo.UserRepository
}

func NewAuthService(cfg config.Config, users repo.UserRepository) AuthService {
	return &authService{cfg: cfg, users: users}
}

func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	_, err := s.users.GetByEmail(ctx, req.Email)
	if err == nil {
		return nil, utils.ErrConflict
	}
	if err != utils.ErrNotFound {
		return nil, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.cfg.BcryptCost)
	if err != nil {
		return nil, utils.WrapDB("bcrypt.generate") // avoid leaking bcrypt internals
	}

	u := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashed),
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}

	token, err := s.signJWT(u.ID, u.Email)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserDTO{ID: u.ID, Name: u.Name, Email: u.Email},
	}, nil
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	u, err := s.users.GetByEmail(ctx, req.Email)
	if err != nil {
		if err == utils.ErrNotFound {
			return nil, utils.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return nil, utils.ErrInvalidCredentials
	}

	token, err := s.signJWT(u.ID, u.Email)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserDTO{ID: u.ID, Name: u.Name, Email: u.Email},
	}, nil
}

type authClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func (s *authService) signJWT(userID uuid.UUID, email string) (string, error) {
	claims := authClaims{
		UserID: userID.String(),
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(s.cfg.JWTSecret))
}

