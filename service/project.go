package service

import (
	"context"

	"github.com/Rakesh2908/taskflow/dto"
	"github.com/Rakesh2908/taskflow/models"
	repo "github.com/Rakesh2908/taskflow/repository"
	"github.com/Rakesh2908/taskflow/utils"
	"github.com/google/uuid"
)

type ProjectService interface {
	Create(ctx context.Context, ownerID uuid.UUID, req dto.CreateProjectRequest) (*models.Project, error)
	List(ctx context.Context, userID uuid.UUID, page *repo.Page) ([]*models.Project, error)
	GetByID(ctx context.Context, projectID uuid.UUID) (*models.Project, error)
	Stats(ctx context.Context, callerID, projectID uuid.UUID) (*dto.ProjectStatsResponse, error)
	Update(ctx context.Context, callerID, projectID uuid.UUID, req dto.UpdateProjectRequest) (*models.Project, error)
	Delete(ctx context.Context, callerID, projectID uuid.UUID) error
}

type projectService struct {
	projects repo.ProjectRepository
	tasks    repo.TaskRepository
}

func NewProjectService(projects repo.ProjectRepository, tasks repo.TaskRepository) ProjectService {
	return &projectService{projects: projects, tasks: tasks}
}

func (s *projectService) Create(ctx context.Context, ownerID uuid.UUID, req dto.CreateProjectRequest) (*models.Project, error) {
	p := &models.Project{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     ownerID,
	}
	if err := s.projects.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *projectService) List(ctx context.Context, userID uuid.UUID, page *repo.Page) ([]*models.Project, error) {
	return s.projects.ListForUser(ctx, userID, page)
}

func (s *projectService) GetByID(ctx context.Context, projectID uuid.UUID) (*models.Project, error) {
	return s.projects.GetByID(ctx, projectID)
}

func (s *projectService) Stats(ctx context.Context, callerID, projectID uuid.UUID) (*dto.ProjectStatsResponse, error) {
	p, err := s.projects.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if p.OwnerID != callerID {
		return nil, utils.ErrForbidden
	}

	rows, err := s.tasks.StatsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	out := &dto.ProjectStatsResponse{
		ProjectID:  projectID,
		ByStatus:   map[string]int{},
		ByAssignee: map[string]map[string]int{},
	}

	for _, row := range rows {
		out.ByStatus[row.Status] += row.Count

		key := "unassigned"
		if row.AssigneeID != nil {
			key = row.AssigneeID.String()
		}
		if out.ByAssignee[key] == nil {
			out.ByAssignee[key] = map[string]int{}
		}
		out.ByAssignee[key][row.Status] += row.Count
	}

	return out, nil
}

func (s *projectService) Update(ctx context.Context, callerID, projectID uuid.UUID, req dto.UpdateProjectRequest) (*models.Project, error) {
	p, err := s.projects.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if p.OwnerID != callerID {
		return nil, utils.ErrForbidden
	}

	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Description != nil {
		p.Description = *req.Description
	}

	if err := s.projects.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *projectService) Delete(ctx context.Context, callerID, projectID uuid.UUID) error {
	p, err := s.projects.GetByID(ctx, projectID)
	if err != nil {
		return err
	}
	if p.OwnerID != callerID {
		return utils.ErrForbidden
	}
	return s.projects.Delete(ctx, projectID)
}

