package service

import (
	"context"

	"github.com/Rakesh2908/taskflow/dto"
	"github.com/Rakesh2908/taskflow/models"
	repo "github.com/Rakesh2908/taskflow/repository"
	"github.com/Rakesh2908/taskflow/utils"
	"github.com/google/uuid"
)

type TaskService interface {
	Create(ctx context.Context, callerID, projectID uuid.UUID, req dto.CreateTaskRequest) (*models.Task, error)
	List(ctx context.Context, projectID uuid.UUID, filters repo.TaskFilters, page *repo.Page) ([]*models.Task, error)
	Update(ctx context.Context, callerID, taskID uuid.UUID, req dto.UpdateTaskRequest) (*models.Task, error)
	Delete(ctx context.Context, callerID, taskID uuid.UUID) error
}

type taskService struct {
	tasks    repo.TaskRepository
	projects repo.ProjectRepository
}

func NewTaskService(tasks repo.TaskRepository, projects repo.ProjectRepository) TaskService {
	return &taskService{tasks: tasks, projects: projects}
}

func (s *taskService) Create(ctx context.Context, callerID, projectID uuid.UUID, req dto.CreateTaskRequest) (*models.Task, error) {
	p, err := s.projects.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if p.OwnerID != callerID {
		return nil, utils.ErrForbidden
	}

	t := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      models.TaskStatusTodo,
		Priority:    models.TaskPriority(req.Priority),
		ProjectID:   projectID,
		AssigneeID:  req.AssigneeID,
		CreatedBy:   callerID,
		DueDate:     req.DueDate,
	}

	if err := s.tasks.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *taskService) List(ctx context.Context, projectID uuid.UUID, filters repo.TaskFilters, page *repo.Page) ([]*models.Task, error) {
	return s.tasks.ListByProject(ctx, projectID, filters, page)
}

func (s *taskService) Update(ctx context.Context, callerID, taskID uuid.UUID, req dto.UpdateTaskRequest) (*models.Task, error) {
	t, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	p, err := s.projects.GetByID(ctx, t.ProjectID)
	if err != nil {
		return nil, err
	}
	if callerID != t.CreatedBy && callerID != p.OwnerID {
		return nil, utils.ErrForbidden
	}

	if req.Title != nil {
		t.Title = *req.Title
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.Status != nil {
		t.Status = models.TaskStatus(*req.Status)
	}
	if req.Priority != nil {
		t.Priority = models.TaskPriority(*req.Priority)
	}
	if req.AssigneeID != nil {
		t.AssigneeID = req.AssigneeID
	}
	if req.DueDate != nil {
		t.DueDate = req.DueDate
	}

	if err := s.tasks.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *taskService) Delete(ctx context.Context, callerID, taskID uuid.UUID) error {
	t, err := s.tasks.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	p, err := s.projects.GetByID(ctx, t.ProjectID)
	if err != nil {
		return err
	}

	if callerID != t.CreatedBy && callerID != p.OwnerID {
		return utils.ErrForbidden
	}

	return s.tasks.Delete(ctx, taskID)
}

