package group

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/encertia/backend/internal/shared"
	"github.com/google/uuid"
)

// Service defines business logic for the group domain.
type Service interface {
	ListGroups(ctx context.Context, actorID uuid.UUID, actorRole string, filters GroupListFilters) (*GroupListResponse, *shared.AppError)
	CreateGroup(ctx context.Context, actorID uuid.UUID, actorRole string, req CreateGroupRequest) (*GroupResponse, *shared.AppError)
	GetGroupByID(ctx context.Context, actorID uuid.UUID, actorRole string, groupID uuid.UUID) (*GroupResponse, *shared.AppError)
	UpdateGroup(ctx context.Context, actorID uuid.UUID, actorRole string, groupID uuid.UUID, req UpdateGroupRequest) (*GroupResponse, *shared.AppError)
	DeleteGroup(ctx context.Context, actorID uuid.UUID, actorRole string, groupID uuid.UUID) *shared.AppError
	ListGroupStudents(ctx context.Context, actorID uuid.UUID, actorRole string, groupID uuid.UUID) (*GroupStudentListResponse, *shared.AppError)
	AddStudentsToGroup(ctx context.Context, actorID uuid.UUID, actorRole string, groupID uuid.UUID, req AddStudentsToGroupRequest) (*GroupStudentListResponse, *shared.AppError)
	RemoveStudentFromGroup(ctx context.Context, actorID uuid.UUID, actorRole string, groupID uuid.UUID, studentID uuid.UUID) *shared.AppError
}

type service struct {
	repo Repository
}

// NewService creates a new group service instance.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ListGroups(ctx context.Context, actorID uuid.UUID, actorRole string, filters GroupListFilters) (*GroupListResponse, *shared.AppError) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.PageSize < 1 {
		filters.PageSize = 10
	}

	items, total, err := s.repo.List(ctx, filters, actorID, actorRole)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(filters.PageSize)))
	}

	return &GroupListResponse{
		Items:      items,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *service) CreateGroup(ctx context.Context, actorID uuid.UUID, actorRole string, req CreateGroupRequest) (*GroupResponse, *shared.AppError) {
	if actorRole != "teacher" && actorRole != "admin" {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "Accés no permès per al rol de l'usuari.")
	}

	trimmedName := strings.TrimSpace(req.Name)
	if trimmedName == "" {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El nom del grup és obligatori.", map[string]interface{}{"field": "name"})
	}

	academicYear := "26-27"
	if req.AcademicYear != nil && strings.TrimSpace(*req.AcademicYear) != "" {
		academicYear = strings.TrimSpace(*req.AcademicYear)
	}

	now := time.Now().UTC()
	g := Group{
		ID:           uuid.New(),
		Name:         trimmedName,
		AcademicYear: academicYear,
		CourseID:     req.CourseID,
		TeacherID:    actorID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(ctx, &g); err != nil {
		return nil, shared.ErrInternal(err)
	}

	created, err := s.repo.GetByID(ctx, g.ID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	return &GroupResponse{Data: *created}, nil
}

func (s *service) GetGroupByID(ctx context.Context, actorID uuid.UUID, actorRole string, groupID uuid.UUID) (*GroupResponse, *shared.AppError) {
	g, err := s.repo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			return nil, shared.ErrNotFound("GROUP_NOT_FOUND", "Grup no trobat.")
		}
		return nil, shared.ErrInternal(err)
	}

	if actorRole == "student" {
		inGroup, err := s.repo.IsStudentInGroup(ctx, groupID, actorID)
		if err != nil {
			return nil, shared.ErrInternal(err)
		}
		if !inGroup {
			return nil, shared.ErrNotFound("GROUP_NOT_FOUND", "Grup no trobat.")
		}
	} else if actorRole == "teacher" {
		if g.TeacherID != actorID {
			return nil, shared.ErrNotFound("GROUP_NOT_FOUND", "Grup no trobat.")
		}
	}

	return &GroupResponse{Data: *g}, nil
}

func (s *service) UpdateGroup(ctx context.Context, actorID uuid.UUID, actorRole string, groupID uuid.UUID, req UpdateGroupRequest) (*GroupResponse, *shared.AppError) {
	if actorRole != "teacher" && actorRole != "admin" {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "Accés no permès per al rol de l'usuari.")
	}

	existing, err := s.repo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			return nil, shared.ErrNotFound("GROUP_NOT_FOUND", "Grup no trobat.")
		}
		return nil, shared.ErrInternal(err)
	}

	if actorRole == "teacher" && existing.TeacherID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No teniu permís per modificar aquest grup.")
	}

	if req.Name != nil {
		trimmedName := strings.TrimSpace(*req.Name)
		if trimmedName == "" {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El nom del grup no pot estar buit.", map[string]interface{}{"field": "name"})
		}
		existing.Name = trimmedName
	}

	if req.AcademicYear != nil {
		trimmedYear := strings.TrimSpace(*req.AcademicYear)
		if trimmedYear != "" {
			existing.AcademicYear = trimmedYear
		}
	}

	if req.CourseID != nil {
		existing.CourseID = req.CourseID
	}

	existing.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, existing); err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			return nil, shared.ErrNotFound("GROUP_NOT_FOUND", "Grup no trobat.")
		}
		return nil, shared.ErrInternal(err)
	}

	updated, err := s.repo.GetByID(ctx, groupID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	return &GroupResponse{Data: *updated}, nil
}

func (s *service) DeleteGroup(ctx context.Context, actorID uuid.UUID, actorRole string, groupID uuid.UUID) *shared.AppError {
	if actorRole != "teacher" && actorRole != "admin" {
		return shared.ErrForbidden(shared.ErrCodeForbidden, "Accés no permès per al rol de l'usuari.")
	}

	existing, err := s.repo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			return shared.ErrNotFound("GROUP_NOT_FOUND", "Grup no trobat.")
		}
		return shared.ErrInternal(err)
	}

	if actorRole == "teacher" && existing.TeacherID != actorID {
		return shared.ErrForbidden(shared.ErrCodeForbidden, "No teniu permís per eliminar aquest grup.")
	}

	if err := s.repo.SoftDelete(ctx, groupID); err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			return shared.ErrNotFound("GROUP_NOT_FOUND", "Grup no trobat.")
		}
		return shared.ErrInternal(err)
	}

	return nil
}

func (s *service) ListGroupStudents(ctx context.Context, actorID uuid.UUID, actorRole string, groupID uuid.UUID) (*GroupStudentListResponse, *shared.AppError) {
	g, err := s.repo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			return nil, shared.ErrNotFound("GROUP_NOT_FOUND", "Grup no trobat.")
		}
		return nil, shared.ErrInternal(err)
	}

	if actorRole == "student" {
		inGroup, err := s.repo.IsStudentInGroup(ctx, groupID, actorID)
		if err != nil {
			return nil, shared.ErrInternal(err)
		}
		if !inGroup {
			return nil, shared.ErrNotFound("GROUP_NOT_FOUND", "Grup no trobat.")
		}
	} else if actorRole == "teacher" {
		if g.TeacherID != actorID {
			return nil, shared.ErrNotFound("GROUP_NOT_FOUND", "Grup no trobat.")
		}
	}

	students, err := s.repo.ListStudents(ctx, groupID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	return &GroupStudentListResponse{
		Items: students,
		Total: len(students),
	}, nil
}

func (s *service) AddStudentsToGroup(ctx context.Context, actorID uuid.UUID, actorRole string, groupID uuid.UUID, req AddStudentsToGroupRequest) (*GroupStudentListResponse, *shared.AppError) {
	if actorRole != "teacher" && actorRole != "admin" {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "Accés no permès per al rol de l'usuari.")
	}

	g, err := s.repo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			return nil, shared.ErrNotFound("GROUP_NOT_FOUND", "Grup no trobat.")
		}
		return nil, shared.ErrInternal(err)
	}

	if actorRole == "teacher" && g.TeacherID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No teniu permís per gestionar aquest grup.")
	}

	if err := s.repo.AddStudents(ctx, groupID, req.StudentIDs); err != nil {
		return nil, shared.ErrInternal(err)
	}

	students, err := s.repo.ListStudents(ctx, groupID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	return &GroupStudentListResponse{
		Items: students,
		Total: len(students),
	}, nil
}

func (s *service) RemoveStudentFromGroup(ctx context.Context, actorID uuid.UUID, actorRole string, groupID uuid.UUID, studentID uuid.UUID) *shared.AppError {
	if actorRole != "teacher" && actorRole != "admin" {
		return shared.ErrForbidden(shared.ErrCodeForbidden, "Accés no permès per al rol de l'usuari.")
	}

	g, err := s.repo.GetByID(ctx, groupID)
	if err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			return shared.ErrNotFound("GROUP_NOT_FOUND", "Grup no trobat.")
		}
		return shared.ErrInternal(err)
	}

	if actorRole == "teacher" && g.TeacherID != actorID {
		return shared.ErrForbidden(shared.ErrCodeForbidden, "No teniu permís per gestionar aquest grup.")
	}

	if err := s.repo.RemoveStudent(ctx, groupID, studentID); err != nil {
		if errors.Is(err, ErrGroupNotFound) {
			return shared.ErrNotFound("STUDENT_NOT_FOUND", "L'alumne no pertany a aquest grup.")
		}
		return shared.ErrInternal(err)
	}

	return nil
}
