package course

import (
	"context"
	"errors"
	"strings"

	"github.com/encertia/backend/internal/shared"
	"github.com/google/uuid"
)

// Service defines application logic interface for course management.
type Service interface {
	ListCourses(ctx context.Context, actorID uuid.UUID, role string, filters CourseListFilters) (*CourseListResponse, *shared.AppError)
	CreateCourse(ctx context.Context, actorID uuid.UUID, role string, req CreateCourseRequest) (*CourseResponse, *shared.AppError)
	GetCourseByID(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID) (*CourseDetailResponse, *shared.AppError)
	UpdateCourse(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID, req UpdateCourseRequest) (*CourseResponse, *shared.AppError)
	DeleteCourse(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID) *shared.AppError

	GetCourseStudents(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID) (*CourseStudentsResponse, *shared.AppError)
	EnrollStudents(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID, req EnrollStudentsRequest) (*CourseStudentsResponse, *shared.AppError)
	UnenrollStudent(ctx context.Context, actorID uuid.UUID, role string, courseID, studentID uuid.UUID) *shared.AppError

	ListCourseUnits(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID) ([]CourseUnitResponse, *shared.AppError)
	CreateCourseUnit(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID, req CreateCourseUnitRequest) (*CourseUnitResponse, *shared.AppError)
	ReorderCourseUnits(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID, unitIDs []uuid.UUID) ([]CourseUnitResponse, *shared.AppError)
	GetCourseUnit(ctx context.Context, actorID uuid.UUID, role string, courseID, unitID uuid.UUID) (*CourseUnitDetailResponse, *shared.AppError)
	UpdateCourseUnit(ctx context.Context, actorID uuid.UUID, role string, courseID, unitID uuid.UUID, req UpdateCourseUnitRequest) (*CourseUnitResponse, *shared.AppError)
	DeleteCourseUnit(ctx context.Context, actorID uuid.UUID, role string, courseID, unitID uuid.UUID) *shared.AppError

	LinkQuizToUnit(ctx context.Context, actorID uuid.UUID, role string, courseID, unitID, quizID uuid.UUID) *shared.AppError
	UnlinkQuizFromUnit(ctx context.Context, actorID uuid.UUID, role string, courseID, unitID, quizID uuid.UUID) *shared.AppError

	GetUnitScript(ctx context.Context, actorID uuid.UUID, role string, courseID, unitID uuid.UUID) ([]ScriptBlockResponse, *shared.AppError)
	UpdateUnitScript(ctx context.Context, actorID uuid.UUID, role string, courseID, unitID uuid.UUID, blocks []CreateScriptBlockRequest) ([]ScriptBlockResponse, *shared.AppError)
}

type service struct {
	repo Repository
}

// NewService creates a new Service instance.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ListCourses(ctx context.Context, actorID uuid.UUID, role string, filters CourseListFilters) (*CourseListResponse, *shared.AppError) {
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 10
	}

	items, total, err := s.repo.ListCourses(ctx, actorID, role, filters)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + filters.PageSize - 1) / filters.PageSize
	}

	responses := make([]CourseResponse, 0, len(items))
	for _, c := range items {
		responses = append(responses, mapToCourseResponse(c))
	}

	return &CourseListResponse{
		Items:      responses,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *service) CreateCourse(ctx context.Context, actorID uuid.UUID, role string, req CreateCourseRequest) (*CourseResponse, *shared.AppError) {
	if role != "admin" && role != "teacher" {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "Només els professors i administradors poden crear cursos.")
	}

	title := strings.TrimSpace(req.Title)
	if len(title) < 3 {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El títol del curs ha de tenir com a mínim 3 caràcters.", map[string]interface{}{"field": "title"})
	}

	code := strings.TrimSpace(req.Code)
	if len(code) < 2 {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El codi del curs ha de tenir com a mínim 2 caràcters.", map[string]interface{}{"field": "code"})
	}

	status := "draft"
	if req.Status != nil && *req.Status != "" {
		st := strings.ToLower(strings.TrimSpace(*req.Status))
		if st != "draft" && st != "active" && st != "archived" {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "L'estat del curs ha de ser 'draft', 'active' o 'archived'.", map[string]interface{}{"field": "status"})
		}
		status = st
	}

	course := &Course{
		Title:       title,
		Code:        code,
		Description: req.Description,
		Status:      status,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		TeacherID:   actorID,
	}

	if err := s.repo.CreateCourse(ctx, course); err != nil {
		if errors.Is(err, ErrCourseCodeConflict) {
			return nil, shared.ErrConflict("COURSE_CODE_ALREADY_EXISTS", "Ja existeix un curs amb aquest codi.")
		}
		return nil, shared.ErrInternal(err)
	}

	// Fetch detail to include teacherName etc.
	created, err := s.repo.GetCourseByID(ctx, course.ID)
	if err != nil {
		resp := mapToCourseResponse(*course)
		return &resp, nil
	}

	resp := mapToCourseResponse(*created)
	return &resp, nil
}

func (s *service) GetCourseByID(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID) (*CourseDetailResponse, *shared.AppError) {
	c, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return nil, shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return nil, shared.ErrInternal(err)
	}

	if role == "teacher" && c.TeacherID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per veure aquest curs.")
	} else if role == "student" {
		enrolled, err := s.repo.IsStudentEnrolled(ctx, courseID, actorID)
		if err != nil {
			return nil, shared.ErrInternal(err)
		}
		if !enrolled {
			return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No estàs matriculat en aquest curs.")
		}
	}

	units, err := s.repo.ListUnitsByCourseID(ctx, courseID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	unitResponses := make([]CourseUnitResponse, 0, len(units))
	for _, u := range units {
		unitResponses = append(unitResponses, mapToCourseUnitResponse(u))
	}

	return &CourseDetailResponse{
		CourseResponse: mapToCourseResponse(*c),
		Units:          unitResponses,
	}, nil
}

func (s *service) UpdateCourse(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID, req UpdateCourseRequest) (*CourseResponse, *shared.AppError) {
	existing, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return nil, shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return nil, shared.ErrInternal(err)
	}

	if role == "teacher" && existing.TeacherID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per modificar aquest curs.")
	}

	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if len(t) < 3 {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El títol del curs ha de tenir com a mínim 3 caràcters.", map[string]interface{}{"field": "title"})
		}
		existing.Title = t
	}

	if req.Code != nil {
		c := strings.TrimSpace(*req.Code)
		if len(c) < 2 {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El codi del curs ha de tenir com a mínim 2 caràcters.", map[string]interface{}{"field": "code"})
		}
		existing.Code = c
	}

	if req.Description != nil {
		existing.Description = req.Description
	}

	if req.Status != nil {
		st := strings.ToLower(strings.TrimSpace(*req.Status))
		if st != "draft" && st != "active" && st != "archived" {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "L'estat del curs ha de ser 'draft', 'active' o 'archived'.", map[string]interface{}{"field": "status"})
		}
		existing.Status = st
	}

	if req.StartDate != nil {
		existing.StartDate = req.StartDate
	}

	if req.EndDate != nil {
		existing.EndDate = req.EndDate
	}

	if err := s.repo.UpdateCourse(ctx, existing); err != nil {
		if errors.Is(err, ErrCourseCodeConflict) {
			return nil, shared.ErrConflict("COURSE_CODE_ALREADY_EXISTS", "Ja existeix un curs amb aquest codi.")
		}
		return nil, shared.ErrInternal(err)
	}

	updated, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		resp := mapToCourseResponse(*existing)
		return &resp, nil
	}

	resp := mapToCourseResponse(*updated)
	return &resp, nil
}

func (s *service) DeleteCourse(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID) *shared.AppError {
	existing, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return shared.ErrInternal(err)
	}

	if role == "teacher" && existing.TeacherID != actorID {
		return shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per eliminar aquest curs.")
	}

	if err := s.repo.DeleteCourse(ctx, courseID); err != nil {
		return shared.ErrInternal(err)
	}

	return nil
}

func (s *service) GetCourseStudents(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID) (*CourseStudentsResponse, *shared.AppError) {
	c, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return nil, shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return nil, shared.ErrInternal(err)
	}

	if role == "teacher" && c.TeacherID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per veure els alumnes d'aquest curs.")
	}

	students, err := s.repo.GetCourseStudents(ctx, courseID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	return &CourseStudentsResponse{
		CourseID: courseID,
		Total:    len(students),
		Students: students,
	}, nil
}

func (s *service) EnrollStudents(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID, req EnrollStudentsRequest) (*CourseStudentsResponse, *shared.AppError) {
	c, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return nil, shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return nil, shared.ErrInternal(err)
	}

	if role == "teacher" && c.TeacherID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per matricular alumnes en aquest curs.")
	}

	if len(req.StudentIDs) == 0 {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "Cal proporcionar almenys un identificador d'alumne.", map[string]interface{}{"field": "studentIds"})
	}

	if err := s.repo.EnrollStudents(ctx, courseID, req.StudentIDs); err != nil {
		return nil, shared.ErrInternal(err)
	}

	return s.GetCourseStudents(ctx, actorID, role, courseID)
}

func (s *service) UnenrollStudent(ctx context.Context, actorID uuid.UUID, role string, courseID, studentID uuid.UUID) *shared.AppError {
	c, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return shared.ErrInternal(err)
	}

	if role == "teacher" && c.TeacherID != actorID {
		return shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per desmatricular alumnes d'aquest curs.")
	}

	if err := s.repo.UnenrollStudent(ctx, courseID, studentID); err != nil {
		return shared.ErrInternal(err)
	}

	return nil
}

func (s *service) ListCourseUnits(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID) ([]CourseUnitResponse, *shared.AppError) {
	c, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return nil, shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return nil, shared.ErrInternal(err)
	}

	if role == "teacher" && c.TeacherID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per consultar les unitats d'aquest curs.")
	} else if role == "student" {
		enrolled, err := s.repo.IsStudentEnrolled(ctx, courseID, actorID)
		if err != nil {
			return nil, shared.ErrInternal(err)
		}
		if !enrolled {
			return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No estàs matriculat en aquest curs.")
		}
	}

	units, err := s.repo.ListUnitsByCourseID(ctx, courseID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	responses := make([]CourseUnitResponse, 0, len(units))
	for _, u := range units {
		responses = append(responses, mapToCourseUnitResponse(u))
	}

	return responses, nil
}

func (s *service) CreateCourseUnit(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID, req CreateCourseUnitRequest) (*CourseUnitResponse, *shared.AppError) {
	c, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return nil, shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return nil, shared.ErrInternal(err)
	}

	if role == "teacher" && c.TeacherID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per crear unitats en aquest curs.")
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El títol de la unitat didàctica és obligatori.", map[string]interface{}{"field": "title"})
	}

	orderIdx := 0
	if req.OrderIndex != nil {
		orderIdx = *req.OrderIndex
	} else {
		existingUnits, _ := s.repo.ListUnitsByCourseID(ctx, courseID)
		orderIdx = len(existingUnits)
	}

	unit := &CourseUnit{
		CourseID:    courseID,
		Title:       title,
		Description: req.Description,
		OrderIndex:  orderIdx,
	}

	if err := s.repo.CreateUnit(ctx, unit); err != nil {
		return nil, shared.ErrInternal(err)
	}

	resp := mapToCourseUnitResponse(*unit)
	return &resp, nil
}

func (s *service) ReorderCourseUnits(ctx context.Context, actorID uuid.UUID, role string, courseID uuid.UUID, unitIDs []uuid.UUID) ([]CourseUnitResponse, *shared.AppError) {
	c, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return nil, shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return nil, shared.ErrInternal(err)
	}

	if role == "teacher" && c.TeacherID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per reordenar les unitats d'aquest curs.")
	}

	if err := s.repo.ReorderUnits(ctx, courseID, unitIDs); err != nil {
		return nil, shared.ErrInternal(err)
	}

	return s.ListCourseUnits(ctx, actorID, role, courseID)
}

func (s *service) GetCourseUnit(ctx context.Context, actorID uuid.UUID, role string, courseID, unitID uuid.UUID) (*CourseUnitDetailResponse, *shared.AppError) {
	c, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return nil, shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return nil, shared.ErrInternal(err)
	}

	if role == "teacher" && c.TeacherID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per veure aquesta unitat.")
	} else if role == "student" {
		enrolled, err := s.repo.IsStudentEnrolled(ctx, courseID, actorID)
		if err != nil {
			return nil, shared.ErrInternal(err)
		}
		if !enrolled {
			return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No estàs matriculat en aquest curs.")
		}
	}

	u, err := s.repo.GetUnitByID(ctx, unitID)
	if err != nil {
		if errors.Is(err, ErrUnitNotFound) {
			return nil, shared.ErrNotFound("UNIT_NOT_FOUND", "No s'ha trobat la unitat didàctica especificada.")
		}
		return nil, shared.ErrInternal(err)
	}

	if u.CourseID != courseID {
		return nil, shared.ErrNotFound("UNIT_NOT_FOUND", "La unitat didàctica no pertany a aquest curs.")
	}

	linkedQuizzes, err := s.repo.GetLinkedQuizzes(ctx, unitID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	blocks, err := s.repo.GetScriptBlocks(ctx, unitID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	blockResponses := make([]ScriptBlockResponse, 0, len(blocks))
	for _, b := range blocks {
		blockResponses = append(blockResponses, mapToScriptBlockResponse(b))
	}

	return &CourseUnitDetailResponse{
		CourseUnitResponse: mapToCourseUnitResponse(*u),
		LinkedQuizzes:      linkedQuizzes,
		ScriptBlocks:       blockResponses,
	}, nil
}

func (s *service) UpdateCourseUnit(ctx context.Context, actorID uuid.UUID, role string, courseID, unitID uuid.UUID, req UpdateCourseUnitRequest) (*CourseUnitResponse, *shared.AppError) {
	c, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return nil, shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return nil, shared.ErrInternal(err)
	}

	if role == "teacher" && c.TeacherID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per modificar aquesta unitat.")
	}

	existing, err := s.repo.GetUnitByID(ctx, unitID)
	if err != nil {
		if errors.Is(err, ErrUnitNotFound) {
			return nil, shared.ErrNotFound("UNIT_NOT_FOUND", "No s'ha trobat la unitat didàctica especificada.")
		}
		return nil, shared.ErrInternal(err)
	}

	if existing.CourseID != courseID {
		return nil, shared.ErrNotFound("UNIT_NOT_FOUND", "La unitat didàctica no pertany a aquest curs.")
	}

	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El títol de la unitat didàctica no pot ser buit.", map[string]interface{}{"field": "title"})
		}
		existing.Title = t
	}

	if req.Description != nil {
		existing.Description = req.Description
	}

	if req.OrderIndex != nil {
		existing.OrderIndex = *req.OrderIndex
	}

	if err := s.repo.UpdateUnit(ctx, existing); err != nil {
		return nil, shared.ErrInternal(err)
	}

	resp := mapToCourseUnitResponse(*existing)
	return &resp, nil
}

func (s *service) DeleteCourseUnit(ctx context.Context, actorID uuid.UUID, role string, courseID, unitID uuid.UUID) *shared.AppError {
	c, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return shared.ErrInternal(err)
	}

	if role == "teacher" && c.TeacherID != actorID {
		return shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per eliminar aquesta unitat.")
	}

	u, err := s.repo.GetUnitByID(ctx, unitID)
	if err != nil {
		if errors.Is(err, ErrUnitNotFound) {
			return shared.ErrNotFound("UNIT_NOT_FOUND", "No s'ha trobat la unitat didàctica especificada.")
		}
		return shared.ErrInternal(err)
	}

	if u.CourseID != courseID {
		return shared.ErrNotFound("UNIT_NOT_FOUND", "La unitat didàctica no pertany a aquest curs.")
	}

	if err := s.repo.DeleteUnit(ctx, unitID); err != nil {
		return shared.ErrInternal(err)
	}

	return nil
}

func (s *service) LinkQuizToUnit(ctx context.Context, actorID uuid.UUID, role string, courseID, unitID, quizID uuid.UUID) *shared.AppError {
	c, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return shared.ErrInternal(err)
	}

	if role == "teacher" && c.TeacherID != actorID {
		return shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per vincular qüestionaris en aquesta unitat.")
	}

	u, err := s.repo.GetUnitByID(ctx, unitID)
	if err != nil {
		if errors.Is(err, ErrUnitNotFound) {
			return shared.ErrNotFound("UNIT_NOT_FOUND", "No s'ha trobat la unitat didàctica especificada.")
		}
		return shared.ErrInternal(err)
	}

	if u.CourseID != courseID {
		return shared.ErrNotFound("UNIT_NOT_FOUND", "La unitat didàctica no pertany a aquest curs.")
	}

	if err := s.repo.LinkQuiz(ctx, unitID, quizID); err != nil {
		return shared.ErrInternal(err)
	}

	return nil
}

func (s *service) UnlinkQuizFromUnit(ctx context.Context, actorID uuid.UUID, role string, courseID, unitID, quizID uuid.UUID) *shared.AppError {
	c, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return shared.ErrInternal(err)
	}

	if role == "teacher" && c.TeacherID != actorID {
		return shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per desvincular qüestionaris d'aquesta unitat.")
	}

	u, err := s.repo.GetUnitByID(ctx, unitID)
	if err != nil {
		if errors.Is(err, ErrUnitNotFound) {
			return shared.ErrNotFound("UNIT_NOT_FOUND", "No s'ha trobat la unitat didàctica especificada.")
		}
		return shared.ErrInternal(err)
	}

	if u.CourseID != courseID {
		return shared.ErrNotFound("UNIT_NOT_FOUND", "La unitat didàctica no pertany a aquest curs.")
	}

	if err := s.repo.UnlinkQuiz(ctx, unitID, quizID); err != nil {
		return shared.ErrInternal(err)
	}

	return nil
}

func (s *service) GetUnitScript(ctx context.Context, actorID uuid.UUID, role string, courseID, unitID uuid.UUID) ([]ScriptBlockResponse, *shared.AppError) {
	c, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return nil, shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return nil, shared.ErrInternal(err)
	}

	if role == "teacher" && c.TeacherID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per consultar el guió d'aquesta unitat.")
	} else if role == "student" {
		enrolled, err := s.repo.IsStudentEnrolled(ctx, courseID, actorID)
		if err != nil {
			return nil, shared.ErrInternal(err)
		}
		if !enrolled {
			return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No estàs matriculat en aquest curs.")
		}
	}

	blocks, err := s.repo.GetScriptBlocks(ctx, unitID)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	responses := make([]ScriptBlockResponse, 0, len(blocks))
	for _, b := range blocks {
		responses = append(responses, mapToScriptBlockResponse(b))
	}

	return responses, nil
}

func (s *service) UpdateUnitScript(ctx context.Context, actorID uuid.UUID, role string, courseID, unitID uuid.UUID, req []CreateScriptBlockRequest) ([]ScriptBlockResponse, *shared.AppError) {
	c, err := s.repo.GetCourseByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, ErrCourseNotFound) {
			return nil, shared.ErrNotFound("COURSE_NOT_FOUND", "No s'ha trobat el curs especificat.")
		}
		return nil, shared.ErrInternal(err)
	}

	if role == "teacher" && c.TeacherID != actorID {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permís per modificar el guió d'aquesta unitat.")
	}

	u, err := s.repo.GetUnitByID(ctx, unitID)
	if err != nil {
		if errors.Is(err, ErrUnitNotFound) {
			return nil, shared.ErrNotFound("UNIT_NOT_FOUND", "No s'ha trobat la unitat didàctica especificada.")
		}
		return nil, shared.ErrInternal(err)
	}

	if u.CourseID != courseID {
		return nil, shared.ErrNotFound("UNIT_NOT_FOUND", "La unitat didàctica no pertany a aquest curs.")
	}

	blocksToInsert := make([]ScriptBlock, 0, len(req))
	for _, r := range req {
		bt := strings.ToLower(strings.TrimSpace(r.BlockType))
		if bt != "material" && bt != "quiz" && bt != "break" {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El tipus de bloc ha de ser 'material', 'quiz' o 'break'.", map[string]interface{}{"field": "blockType"})
		}
		title := strings.TrimSpace(r.Title)
		if title == "" {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El títol del bloc de guió és obligatori.", map[string]interface{}{"field": "title"})
		}

		blocksToInsert = append(blocksToInsert, ScriptBlock{
			UnitID:          unitID,
			BlockType:       bt,
			OrderIndex:      r.OrderIndex,
			Title:           title,
			Description:     r.Description,
			MaterialID:      r.MaterialID,
			PdfURL:          r.PdfURL,
			StartPage:       r.StartPage,
			EndPage:         r.EndPage,
			QuizID:          r.QuizID,
			DurationMinutes: r.DurationMinutes,
		})
	}

	updatedBlocks, err := s.repo.ReplaceScriptBlocks(ctx, unitID, blocksToInsert)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	responses := make([]ScriptBlockResponse, 0, len(updatedBlocks))
	for _, b := range updatedBlocks {
		responses = append(responses, mapToScriptBlockResponse(b))
	}

	return responses, nil
}

func mapToCourseResponse(c Course) CourseResponse {
	return CourseResponse{
		ID:                    c.ID,
		Title:                 c.Title,
		Code:                  c.Code,
		Description:           c.Description,
		Status:                c.Status,
		StartDate:             c.StartDate,
		EndDate:               c.EndDate,
		TeacherID:             c.TeacherID,
		TeacherName:           c.TeacherName,
		EnrolledStudentsCount: c.EnrolledStudentsCount,
		UnitsCount:            c.UnitsCount,
		CreatedAt:             c.CreatedAt,
		UpdatedAt:             c.UpdatedAt,
	}
}

func mapToCourseUnitResponse(u CourseUnit) CourseUnitResponse {
	return CourseUnitResponse{
		ID:           u.ID,
		CourseID:     u.CourseID,
		Title:        u.Title,
		Description:  u.Description,
		OrderIndex:   u.OrderIndex,
		QuizzesCount: u.QuizzesCount,
		BlocksCount:  u.BlocksCount,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

func mapToScriptBlockResponse(b ScriptBlock) ScriptBlockResponse {
	return ScriptBlockResponse{
		ID:              b.ID,
		UnitID:          b.UnitID,
		BlockType:       b.BlockType,
		OrderIndex:      b.OrderIndex,
		Title:           b.Title,
		Description:     b.Description,
		MaterialID:      b.MaterialID,
		PdfURL:          b.PdfURL,
		StartPage:       b.StartPage,
		EndPage:         b.EndPage,
		QuizID:          b.QuizID,
		QuizTitle:       b.QuizTitle,
		DurationMinutes: b.DurationMinutes,
		CreatedAt:       b.CreatedAt,
	}
}
