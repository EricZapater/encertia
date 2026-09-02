package course

import (
	"context"

	"github.com/google/uuid"
)

type mockRepository struct {
	courses        map[uuid.UUID]*Course
	enrollments    map[uuid.UUID][]uuid.UUID // courseID -> []studentID
	units          map[uuid.UUID]*CourseUnit
	unitQuizzes    map[uuid.UUID][]LinkedQuiz
	scriptBlocks   map[uuid.UUID][]ScriptBlock
	mockStudents   map[uuid.UUID]EnrolledStudentResponse
	codeConflict   bool
	forceErr       error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		courses:      make(map[uuid.UUID]*Course),
		enrollments:  make(map[uuid.UUID][]uuid.UUID),
		units:        make(map[uuid.UUID]*CourseUnit),
		unitQuizzes:  make(map[uuid.UUID][]LinkedQuiz),
		scriptBlocks: make(map[uuid.UUID][]ScriptBlock),
		mockStudents: make(map[uuid.UUID]EnrolledStudentResponse),
	}
}

func (m *mockRepository) CreateCourse(ctx context.Context, c *Course) error {
	if m.forceErr != nil {
		return m.forceErr
	}
	if m.codeConflict {
		return ErrCourseCodeConflict
	}
	for _, existing := range m.courses {
		if existing.DeletedAt == nil && existing.Code == c.Code && existing.ID != c.ID {
			return ErrCourseCodeConflict
		}
	}
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	m.courses[c.ID] = c
	return nil
}

func (m *mockRepository) GetCourseByID(ctx context.Context, id uuid.UUID) (*Course, error) {
	if m.forceErr != nil {
		return nil, m.forceErr
	}
	c, exists := m.courses[id]
	if !exists || c.DeletedAt != nil {
		return nil, ErrCourseNotFound
	}
	return c, nil
}

func (m *mockRepository) GetCourseByCode(ctx context.Context, code string) (*Course, error) {
	if m.forceErr != nil {
		return nil, m.forceErr
	}
	for _, c := range m.courses {
		if c.DeletedAt == nil && c.Code == code {
			return c, nil
		}
	}
	return nil, ErrCourseNotFound
}

func (m *mockRepository) UpdateCourse(ctx context.Context, c *Course) error {
	if m.forceErr != nil {
		return m.forceErr
	}
	existing, exists := m.courses[c.ID]
	if !exists || existing.DeletedAt != nil {
		return ErrCourseNotFound
	}
	m.courses[c.ID] = c
	return nil
}

func (m *mockRepository) DeleteCourse(ctx context.Context, id uuid.UUID) error {
	if m.forceErr != nil {
		return m.forceErr
	}
	c, exists := m.courses[id]
	if !exists || c.DeletedAt != nil {
		return ErrCourseNotFound
	}
	now := c.CreatedAt
	c.DeletedAt = &now
	return nil
}

func (m *mockRepository) ListCourses(ctx context.Context, actorID uuid.UUID, role string, filters CourseListFilters) ([]Course, int, error) {
	if m.forceErr != nil {
		return nil, 0, m.forceErr
	}
	res := make([]Course, 0)
	for _, c := range m.courses {
		if c.DeletedAt != nil {
			continue
		}
		if role == "teacher" && c.TeacherID != actorID {
			continue
		}
		if role == "student" {
			enrolled := false
			for _, sID := range m.enrollments[c.ID] {
				if sID == actorID {
					enrolled = true
					break
				}
			}
			if !enrolled {
				continue
			}
		}
		if filters.Status != "" && c.Status != filters.Status {
			continue
		}
		res = append(res, *c)
	}
	return res, len(res), nil
}

func (m *mockRepository) EnrollStudents(ctx context.Context, courseID uuid.UUID, studentIDs []uuid.UUID) error {
	if m.forceErr != nil {
		return m.forceErr
	}
	current := m.enrollments[courseID]
	for _, sID := range studentIDs {
		already := false
		for _, existing := range current {
			if existing == sID {
				already = true
				break
			}
		}
		if !already {
			current = append(current, sID)
		}
	}
	m.enrollments[courseID] = current
	return nil
}

func (m *mockRepository) UnenrollStudent(ctx context.Context, courseID, studentID uuid.UUID) error {
	if m.forceErr != nil {
		return m.forceErr
	}
	current := m.enrollments[courseID]
	updated := make([]uuid.UUID, 0)
	found := false
	for _, sID := range current {
		if sID == studentID {
			found = true
			continue
		}
		updated = append(updated, sID)
	}
	if !found {
		return ErrCourseNotFound
	}
	m.enrollments[courseID] = updated
	return nil
}

func (m *mockRepository) GetCourseStudents(ctx context.Context, courseID uuid.UUID) ([]EnrolledStudentResponse, error) {
	if m.forceErr != nil {
		return nil, m.forceErr
	}
	res := make([]EnrolledStudentResponse, 0)
	for _, sID := range m.enrollments[courseID] {
		if st, ok := m.mockStudents[sID]; ok {
			res = append(res, st)
		} else {
			res = append(res, EnrolledStudentResponse{
				ID:        sID,
				FirstName: "Student",
				LastName:  "Test",
				Email:     "student@example.com",
			})
		}
	}
	return res, nil
}

func (m *mockRepository) IsStudentEnrolled(ctx context.Context, courseID, studentID uuid.UUID) (bool, error) {
	if m.forceErr != nil {
		return false, m.forceErr
	}
	for _, sID := range m.enrollments[courseID] {
		if sID == studentID {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockRepository) CreateUnit(ctx context.Context, unit *CourseUnit) error {
	if m.forceErr != nil {
		return m.forceErr
	}
	if unit.ID == uuid.Nil {
		unit.ID = uuid.New()
	}
	m.units[unit.ID] = unit
	return nil
}

func (m *mockRepository) GetUnitByID(ctx context.Context, unitID uuid.UUID) (*CourseUnit, error) {
	if m.forceErr != nil {
		return nil, m.forceErr
	}
	u, exists := m.units[unitID]
	if !exists || u.DeletedAt != nil {
		return nil, ErrUnitNotFound
	}
	return u, nil
}

func (m *mockRepository) ListUnitsByCourseID(ctx context.Context, courseID uuid.UUID) ([]CourseUnit, error) {
	if m.forceErr != nil {
		return nil, m.forceErr
	}
	res := make([]CourseUnit, 0)
	for _, u := range m.units {
		if u.CourseID == courseID && u.DeletedAt == nil {
			res = append(res, *u)
		}
	}
	return res, nil
}

func (m *mockRepository) UpdateUnit(ctx context.Context, unit *CourseUnit) error {
	if m.forceErr != nil {
		return m.forceErr
	}
	existing, exists := m.units[unit.ID]
	if !exists || existing.DeletedAt != nil {
		return ErrUnitNotFound
	}
	m.units[unit.ID] = unit
	return nil
}

func (m *mockRepository) DeleteUnit(ctx context.Context, unitID uuid.UUID) error {
	if m.forceErr != nil {
		return m.forceErr
	}
	u, exists := m.units[unitID]
	if !exists || u.DeletedAt != nil {
		return ErrUnitNotFound
	}
	now := u.CreatedAt
	u.DeletedAt = &now
	return nil
}

func (m *mockRepository) ReorderUnits(ctx context.Context, courseID uuid.UUID, unitIDs []uuid.UUID) error {
	if m.forceErr != nil {
		return m.forceErr
	}
	for idx, uID := range unitIDs {
		if u, exists := m.units[uID]; exists && u.CourseID == courseID {
			u.OrderIndex = idx
		}
	}
	return nil
}

func (m *mockRepository) LinkQuiz(ctx context.Context, unitID, quizID uuid.UUID) error {
	if m.forceErr != nil {
		return m.forceErr
	}
	current := m.unitQuizzes[unitID]
	for _, q := range current {
		if q.ID == quizID {
			return nil
		}
	}
	m.unitQuizzes[unitID] = append(current, LinkedQuiz{
		ID:             quizID,
		Title:          "Sample Quiz",
		QuestionsCount: 5,
	})
	return nil
}

func (m *mockRepository) UnlinkQuiz(ctx context.Context, unitID, quizID uuid.UUID) error {
	if m.forceErr != nil {
		return m.forceErr
	}
	current := m.unitQuizzes[unitID]
	updated := make([]LinkedQuiz, 0)
	for _, q := range current {
		if q.ID != quizID {
			updated = append(updated, q)
		}
	}
	m.unitQuizzes[unitID] = updated
	return nil
}

func (m *mockRepository) GetLinkedQuizzes(ctx context.Context, unitID uuid.UUID) ([]LinkedQuiz, error) {
	if m.forceErr != nil {
		return nil, m.forceErr
	}
	return m.unitQuizzes[unitID], nil
}

func (m *mockRepository) GetScriptBlocks(ctx context.Context, unitID uuid.UUID) ([]ScriptBlock, error) {
	if m.forceErr != nil {
		return nil, m.forceErr
	}
	return m.scriptBlocks[unitID], nil
}

func (m *mockRepository) ReplaceScriptBlocks(ctx context.Context, unitID uuid.UUID, blocks []ScriptBlock) ([]ScriptBlock, error) {
	if m.forceErr != nil {
		return nil, m.forceErr
	}
	for i := range blocks {
		if blocks[i].ID == uuid.Nil {
			blocks[i].ID = uuid.New()
		}
		blocks[i].UnitID = unitID
	}
	m.scriptBlocks[unitID] = blocks
	return blocks, nil
}
