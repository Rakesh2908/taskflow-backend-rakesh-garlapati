package service_test

import (
	"context"
	"strings"
	"testing"

	"github.com/Rakesh2908/taskflow/config"
	"github.com/Rakesh2908/taskflow/dto"
	"github.com/Rakesh2908/taskflow/models"
	repo "github.com/Rakesh2908/taskflow/repository"
	"github.com/Rakesh2908/taskflow/service"
	"github.com/Rakesh2908/taskflow/utils"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	getByEmailFn func(ctx context.Context, email string) (*models.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, u *models.User) error {
	return nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.getByEmailFn(ctx, email)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	return nil, utils.ErrNotFound
}

var _ repo.UserRepository = (*mockUserRepo)(nil)

func TestAuthServiceLogin_UserNotFound(t *testing.T) {
	svc := service.NewAuthService(
		config.Config{JWTSecret: "test", BcryptCost: 12},
		&mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
				return nil, utils.ErrNotFound
			},
		},
	)

	_, err := svc.Login(context.Background(), dto.LoginRequest{Email: "x@example.com", Password: "password123"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != utils.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthServiceLogin_WrongPassword(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}

	u := &models.User{
		ID:       uuid.New(),
		Name:     "User",
		Email:    "x@example.com",
		Password: string(hash),
	}

	svc := service.NewAuthService(
		config.Config{JWTSecret: "test", BcryptCost: 12},
		&mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
				return u, nil
			},
		},
	)

	_, err = svc.Login(context.Background(), dto.LoginRequest{Email: u.Email, Password: "wrong"})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != utils.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthServiceLogin_ValidCredentials(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	if err != nil {
		t.Fatalf("bcrypt.GenerateFromPassword: %v", err)
	}

	u := &models.User{
		ID:       uuid.New(),
		Name:     "User",
		Email:    "x@example.com",
		Password: string(hash),
	}

	svc := service.NewAuthService(
		config.Config{JWTSecret: "test", BcryptCost: 12},
		&mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
				return u, nil
			},
		},
	)

	res, err := svc.Login(context.Background(), dto.LoginRequest{Email: u.Email, Password: "password123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || res.Token == "" {
		t.Fatalf("expected non-empty token")
	}
	if strings.Count(res.Token, ".") != 2 {
		t.Fatalf("expected jwt-like token, got %q", res.Token)
	}
}

