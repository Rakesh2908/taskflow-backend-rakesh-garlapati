package repository

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"

	"github.com/Rakesh2908/taskflow/models"
	"github.com/Rakesh2908/taskflow/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository interface {
	Create(ctx context.Context, t *models.Task) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Task, error)
	ListByProject(ctx context.Context, projectID uuid.UUID, filters TaskFilters, page *Page) ([]*models.Task, error)
	StatsByProject(ctx context.Context, projectID uuid.UUID) ([]TaskStatsRow, error)
	Update(ctx context.Context, t *models.Task) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type TaskFilters struct {
	Status     *string
	AssigneeID *uuid.UUID
}

type TaskStatsRow struct {
	Status     string
	AssigneeID *uuid.UUID
	Count      int
}

type postgresTaskRepo struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) TaskRepository {
	return &postgresTaskRepo{pool: pool}
}

func (r *postgresTaskRepo) Create(ctx context.Context, t *models.Task) error {
	const op = "task.create"
	const q = `
INSERT INTO tasks (
  title, description, status, priority,
  project_id, assignee_id, created_by, due_date
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, created_at, updated_at
`

	err := r.pool.QueryRow(
		ctx,
		q,
		t.Title,
		t.Description,
		string(t.Status),
		string(t.Priority),
		t.ProjectID,
		t.AssigneeID,
		t.CreatedBy,
		t.DueDate,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		slog.Error("db query failed", "op", op, "err", err)
		return utils.WrapDB(op)
	}
	return nil
}

func (r *postgresTaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	const op = "task.get_by_id"
	const q = `
SELECT id, title, description, status, priority,
       project_id, assignee_id, created_by, due_date,
       created_at, updated_at
FROM tasks
WHERE id = $1
`

	var t models.Task
	var status string
	var priority string
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&status,
		&priority,
		&t.ProjectID,
		&t.AssigneeID,
		&t.CreatedBy,
		&t.DueDate,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrNotFound
		}
		slog.Error("db query failed", "op", op, "err", err)
		return nil, utils.WrapDB(op)
	}
	t.Status = models.TaskStatus(status)
	t.Priority = models.TaskPriority(priority)
	return &t, nil
}

func (r *postgresTaskRepo) ListByProject(ctx context.Context, projectID uuid.UUID, filters TaskFilters, page *Page) ([]*models.Task, error) {
	const op = "task.list_by_project"

	var b strings.Builder
	b.WriteString(`
SELECT id, title, description, status, priority,
       project_id, assignee_id, created_by, due_date,
       created_at, updated_at
FROM tasks
WHERE project_id = $1
`)

	args := make([]any, 0, 3)
	args = append(args, projectID)
	param := 2

	if filters.Status != nil {
		b.WriteString("  AND status = $")
		b.WriteString(strconv.Itoa(param))
		b.WriteString("\n")
		args = append(args, *filters.Status)
		param++
	}

	if filters.AssigneeID != nil {
		b.WriteString("  AND assignee_id = $")
		b.WriteString(strconv.Itoa(param))
		b.WriteString("\n")
		args = append(args, *filters.AssigneeID)
		param++
	}

	b.WriteString("ORDER BY created_at DESC\n")
	if page != nil {
		b.WriteString("LIMIT $")
		b.WriteString(strconv.Itoa(param))
		b.WriteString(" OFFSET $")
		b.WriteString(strconv.Itoa(param + 1))
		b.WriteString("\n")
		args = append(args, page.Limit, page.Offset)
	}

	rows, err := r.pool.Query(ctx, b.String(), args...)
	if err != nil {
		slog.Error("db query failed", "op", op, "err", err)
		return nil, utils.WrapDB(op)
	}
	defer rows.Close()

	out := make([]*models.Task, 0, 32)
	for rows.Next() {
		var t models.Task
		var status string
		var priority string
		if err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Description,
			&status,
			&priority,
			&t.ProjectID,
			&t.AssigneeID,
			&t.CreatedBy,
			&t.DueDate,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			slog.Error("db scan failed", "op", op, "err", err)
			return nil, utils.WrapDB(op)
		}
		t.Status = models.TaskStatus(status)
		t.Priority = models.TaskPriority(priority)
		out = append(out, &t)
	}
	if err := rows.Err(); err != nil {
		slog.Error("db rows failed", "op", op, "err", err)
		return nil, utils.WrapDB(op)
	}
	return out, nil
}

func (r *postgresTaskRepo) StatsByProject(ctx context.Context, projectID uuid.UUID) ([]TaskStatsRow, error) {
	const op = "task.stats_by_project"
	const q = `
SELECT status, assignee_id, COUNT(*)::int
FROM tasks
WHERE project_id = $1
GROUP BY status, assignee_id
`

	rows, err := r.pool.Query(ctx, q, projectID)
	if err != nil {
		slog.Error("db query failed", "op", op, "err", err)
		return nil, utils.WrapDB(op)
	}
	defer rows.Close()

	out := make([]TaskStatsRow, 0, 8)
	for rows.Next() {
		var row TaskStatsRow
		if err := rows.Scan(&row.Status, &row.AssigneeID, &row.Count); err != nil {
			slog.Error("db scan failed", "op", op, "err", err)
			return nil, utils.WrapDB(op)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		slog.Error("db rows failed", "op", op, "err", err)
		return nil, utils.WrapDB(op)
	}
	return out, nil
}

func (r *postgresTaskRepo) Update(ctx context.Context, t *models.Task) error {
	const op = "task.update"
	const q = `
UPDATE tasks
SET title = $2,
    description = $3,
    status = $4,
    priority = $5,
    project_id = $6,
    assignee_id = $7,
    created_by = $8,
    due_date = $9
WHERE id = $1
RETURNING updated_at
`

	err := r.pool.QueryRow(
		ctx,
		q,
		t.ID,
		t.Title,
		t.Description,
		string(t.Status),
		string(t.Priority),
		t.ProjectID,
		t.AssigneeID,
		t.CreatedBy,
		t.DueDate,
	).Scan(&t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return utils.ErrNotFound
		}
		slog.Error("db query failed", "op", op, "err", err)
		return utils.WrapDB(op)
	}

	return nil
}

func (r *postgresTaskRepo) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "task.delete"
	const q = `DELETE FROM tasks WHERE id = $1`

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

