package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/Rakesh2908/taskflow/models"
	"github.com/Rakesh2908/taskflow/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	Create(ctx context.Context, u *models.User) error
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
}

type postgresUserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &postgresUserRepo{pool: pool}
}

func (r *postgresUserRepo) Create(ctx context.Context, u *models.User) error {
	const op = "user.create"
	const q = `
INSERT INTO users (name, email, password)
VALUES ($1, $2, $3)
RETURNING id, created_at
`

	err := r.pool.QueryRow(ctx, q, u.Name, u.Email, u.Password).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		slog.Error("db query failed", "op", op, "err", err)
		return utils.WrapDB(op)
	}
	return nil
}

func (r *postgresUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	const op = "user.get_by_email"
	const q = `
SELECT id, name, email, password, created_at
FROM users
WHERE email = $1
`

	var u models.User
	err := r.pool.QueryRow(ctx, q, email).Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrNotFound
		}
		slog.Error("db query failed", "op", op, "err", err)
		return nil, utils.WrapDB(op)
	}
	return &u, nil
}

func (r *postgresUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	const op = "user.get_by_id"
	const q = `
SELECT id, name, email, password, created_at
FROM users
WHERE id = $1
`

	var u models.User
	err := r.pool.QueryRow(ctx, q, id).Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrNotFound
		}
		slog.Error("db query failed", "op", op, "err", err)
		return nil, utils.WrapDB(op)
	}
	return &u, nil
}

