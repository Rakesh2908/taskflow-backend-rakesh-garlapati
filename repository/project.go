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

type ProjectRepository interface {
	Create(ctx context.Context, p *models.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error)
	ListForUser(ctx context.Context, userID uuid.UUID, page *Page) ([]*models.Project, error)
	Update(ctx context.Context, p *models.Project) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type postgresProjectRepo struct {
	pool *pgxpool.Pool
}

func NewProjectRepository(pool *pgxpool.Pool) ProjectRepository {
	return &postgresProjectRepo{pool: pool}
}

func (r *postgresProjectRepo) Create(ctx context.Context, p *models.Project) error {
	const op = "project.create"
	const q = `
INSERT INTO projects (name, description, owner_id)
VALUES ($1, $2, $3)
RETURNING id, created_at
`

	err := r.pool.QueryRow(ctx, q, p.Name, p.Description, p.OwnerID).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		slog.Error("db query failed", "op", op, "err", err)
		return utils.WrapDB(op)
	}
	return nil
}

func (r *postgresProjectRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	const op = "project.get_by_id"
	const q = `
SELECT id, name, description, owner_id, created_at
FROM projects
WHERE id = $1
`

	var p models.Project
	err := r.pool.QueryRow(ctx, q, id).Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrNotFound
		}
		slog.Error("db query failed", "op", op, "err", err)
		return nil, utils.WrapDB(op)
	}
	return &p, nil
}

func (r *postgresProjectRepo) ListForUser(ctx context.Context, userID uuid.UUID, page *Page) ([]*models.Project, error) {
	const op = "project.list_for_user"
	q := `
SELECT DISTINCT p.id, p.name, p.description, p.owner_id, p.created_at
FROM projects p
WHERE p.owner_id = $1
   OR EXISTS (
     SELECT 1
     FROM tasks t
     WHERE t.project_id = p.id
       AND t.assignee_id = $1
   )
ORDER BY p.created_at DESC
`

	args := []any{userID}
	if page != nil {
		q += "\nLIMIT $2 OFFSET $3\n"
		args = append(args, page.Limit, page.Offset)
	}

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		slog.Error("db query failed", "op", op, "err", err)
		return nil, utils.WrapDB(op)
	}
	defer rows.Close()

	out := make([]*models.Project, 0, 16)
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.OwnerID, &p.CreatedAt); err != nil {
			slog.Error("db scan failed", "op", op, "err", err)
			return nil, utils.WrapDB(op)
		}
		out = append(out, &p)
	}
	if err := rows.Err(); err != nil {
		slog.Error("db rows failed", "op", op, "err", err)
		return nil, utils.WrapDB(op)
	}
	return out, nil
}

func (r *postgresProjectRepo) Update(ctx context.Context, p *models.Project) error {
	const op = "project.update"
	const q = `
UPDATE projects
SET name = $2,
    description = $3,
    owner_id = $4
WHERE id = $1
`

	ct, err := r.pool.Exec(ctx, q, p.ID, p.Name, p.Description, p.OwnerID)
	if err != nil {
		slog.Error("db exec failed", "op", op, "err", err)
		return utils.WrapDB(op)
	}
	if ct.RowsAffected() == 0 {
		return utils.ErrNotFound
	}
	return nil
}

func (r *postgresProjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "project.delete"
	const q = `DELETE FROM projects WHERE id = $1`

	ct, err := r.pool.Exec(ctx, q, id)
	if err != nil {
		slog.Error("db exec failed", "op", op, "err", err)
		return utils.WrapDB(op)
	}
	if ct.RowsAffected() == 0 {
		return utils.ErrNotFound
	}
	return nil
}

