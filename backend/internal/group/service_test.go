package group

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type mockRepository struct {
	groups        map[uuid.UUID]*Group
	groupStudents map[uuid.UUID][]uuid.UUID
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		groups:        make(map[uuid.UUID]*Group),
		groupStudents: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *mockRepository) Create(ctx context.Context, g *Group) error {
	m.groups[g.ID] = g
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uuid.UUID) (*Group, error) {
	g, ok := m.groups[id]
	if !ok || g.DeletedAt != nil {
		return nil, ErrGroupNotFound
	}
	return g, nil
}

func (m *mockRepository) Update(ctx context.Context, g *Group) error {
	if _, ok := m.groups[g.ID]; !ok {
		return ErrGroupNotFound
	}
	m.groups[g.ID] = g
	return nil
}

func (m *mockRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	g, ok := m.groups[id]
	if !ok || g.DeletedAt != nil {
		return ErrGroupNotFound
	}
	now := g.UpdatedAt
	g.DeletedAt = &now
	return nil
}

func (m *mockRepository) List(ctx context.Context, filters GroupListFilters, actorID uuid.UUID, actorRole string) ([]Group, int, error) {
	res := make([]Group, 0)
	for _, g := range m.groups {
		if g.DeletedAt != nil {
			continue
		}
		if actorRole == "teacher" && g.TeacherID != actorID {
			continue
		}
		res = append(res, *g)
	}
	return res, len(res), nil
}

func (m *mockRepository) ListStudents(ctx context.Context, groupID uuid.UUID) ([]GroupStudent, error) {
	sIDs := m.groupStudents[groupID]
	res := make([]GroupStudent, len(sIDs))
	for i, id := range sIDs {
		res[i] = GroupStudent{
			ID:       id,
			Email:    "student@test.com",
			FullName: "Test Student",
		}
	}
	return res, nil
}

func (m *mockRepository) AddStudents(ctx context.Context, groupID uuid.UUID, studentIDs []uuid.UUID) error {
	m.groupStudents[groupID] = append(m.groupStudents[groupID], studentIDs...)
	return nil
}

func (m *mockRepository) RemoveStudent(ctx context.Context, groupID uuid.UUID, studentID uuid.UUID) error {
	sIDs := m.groupStudents[groupID]
	newIDs := make([]uuid.UUID, 0)
	found := false
	for _, id := range sIDs {
		if id == studentID {
			found = true
			continue
		}
		newIDs = append(newIDs, id)
	}
	if !found {
		return ErrGroupNotFound
	}
	m.groupStudents[groupID] = newIDs
	return nil
}

func (m *mockRepository) IsStudentInGroup(ctx context.Context, groupID uuid.UUID, studentID uuid.UUID) (bool, error) {
	sIDs := m.groupStudents[groupID]
	for _, id := range sIDs {
		if id == studentID {
			return true, nil
		}
	}
	return false, nil
}

func TestCreateGroup(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	teacherID := uuid.New()
	ctx := context.Background()

	// 1. Creation with empty name should fail
	_, appErr := svc.CreateGroup(ctx, teacherID, "teacher", CreateGroupRequest{
		Name: "",
	})
	if appErr == nil {
		t.Errorf("S'esperava un error en crear un grup sense nom, però ha tingut èxit")
	}

	// 2. Creation with valid name and default academic year "26-27"
	res, appErr := svc.CreateGroup(ctx, teacherID, "teacher", CreateGroupRequest{
		Name: "Grup A",
	})
	if appErr != nil {
		t.Fatalf("Error inesperat creant el grup: %v", appErr)
	}
	if res.Data.Name != "Grup A" {
		t.Errorf("Nom esperat 'Grup A', obtingut '%s'", res.Data.Name)
	}
	if res.Data.AcademicYear != "26-27" {
		t.Errorf("Curs acadèmic per defecte esperat '26-27', obtingut '%s'", res.Data.AcademicYear)
	}
}

func TestUpdateAndDeleteGroup(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	teacherID := uuid.New()
	ctx := context.Background()

	// Create group
	res, appErr := svc.CreateGroup(ctx, teacherID, "teacher", CreateGroupRequest{
		Name: "Grup B&F",
	})
	if appErr != nil {
		t.Fatalf("Error creant el grup: %v", appErr)
	}
	groupID := res.Data.ID

	// Update group
	newName := "Grup B&F Segon Curs"
	updated, appErr := svc.UpdateGroup(ctx, teacherID, "teacher", groupID, UpdateGroupRequest{
		Name: &newName,
	})
	if appErr != nil {
		t.Fatalf("Error actualitzant el grup: %v", appErr)
	}
	if updated.Data.Name != newName {
		t.Errorf("Esperat nom '%s', obtingut '%s'", newName, updated.Data.Name)
	}

	// Soft Delete group
	appErr = svc.DeleteGroup(ctx, teacherID, "teacher", groupID)
	if appErr != nil {
		t.Fatalf("Error esborrant el grup: %v", appErr)
	}

	// Fetch should fail after soft delete
	_, appErr = svc.GetGroupByID(ctx, teacherID, "teacher", groupID)
	if appErr == nil {
		t.Errorf("S'esperava un 404 en cercar un grup esborrat")
	}
}

func TestAddAndRemoveStudents(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	teacherID := uuid.New()
	studentID := uuid.New()
	ctx := context.Background()

	// Create group
	res, appErr := svc.CreateGroup(ctx, teacherID, "teacher", CreateGroupRequest{
		Name: "Grup 1",
	})
	if appErr != nil {
		t.Fatalf("Error creant el grup: %v", appErr)
	}
	groupID := res.Data.ID

	// Add student
	studentsRes, appErr := svc.AddStudentsToGroup(ctx, teacherID, "teacher", groupID, AddStudentsToGroupRequest{
		StudentIDs: []uuid.UUID{studentID},
	})
	if appErr != nil {
		t.Fatalf("Error afegint alumne: %v", appErr)
	}
	if studentsRes.Total != 1 {
		t.Errorf("S'esperava 1 alumne al grup, s'han trobat %d", studentsRes.Total)
	}

	// Remove student
	appErr = svc.RemoveStudentFromGroup(ctx, teacherID, "teacher", groupID, studentID)
	if appErr != nil {
		t.Fatalf("Error eliminant alumne del grup: %v", appErr)
	}
}
