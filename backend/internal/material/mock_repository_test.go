package material

import (
	"context"

	"github.com/google/uuid"
)

type mockRepository struct {
	materials           map[uuid.UUID]*Material
	unitMaterials       map[uuid.UUID][]uuid.UUID
	unitMaterialOrders  map[string]int
	materialViews       map[uuid.UUID][]uuid.UUID
	courseTeachers      map[uuid.UUID]uuid.UUID
	unitCourses         map[uuid.UUID]uuid.UUID
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		materials:          make(map[uuid.UUID]*Material),
		unitMaterials:      make(map[uuid.UUID][]uuid.UUID),
		unitMaterialOrders: make(map[string]int),
		materialViews:      make(map[uuid.UUID][]uuid.UUID),
		courseTeachers:     make(map[uuid.UUID]uuid.UUID),
		unitCourses:        make(map[uuid.UUID]uuid.UUID),
	}
}

func (m *mockRepository) Create(ctx context.Context, mat *Material) error {
	if mat.ID == uuid.Nil {
		mat.ID = uuid.New()
	}
	m.materials[mat.ID] = mat
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uuid.UUID) (*Material, error) {
	mat, ok := m.materials[id]
	if !ok || mat.DeletedAt != nil {
		return nil, ErrNotFound
	}
	return mat, nil
}

func (m *mockRepository) Update(ctx context.Context, mat *Material) error {
	existing, ok := m.materials[mat.ID]
	if !ok || existing.DeletedAt != nil {
		return ErrNotFound
	}
	m.materials[mat.ID] = mat
	return nil
}

func (m *mockRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	mat, ok := m.materials[id]
	if !ok || mat.DeletedAt != nil {
		return ErrNotFound
	}
	now := ctx.Value("now")
	_ = now
	mat.DeletedAt = &mat.CreatedAt
	return nil
}

func (m *mockRepository) List(ctx context.Context, filters MaterialListFilters, teacherID *uuid.UUID) ([]Material, int, error) {
	var result []Material
	for _, mat := range m.materials {
		if mat.DeletedAt != nil {
			continue
		}
		if teacherID != nil && mat.TeacherID != *teacherID {
			continue
		}
		if filters.MaterialType != "" && string(mat.MaterialType) != filters.MaterialType {
			continue
		}
		result = append(result, *mat)
	}
	return result, len(result), nil
}

func (m *mockRepository) RecordView(ctx context.Context, materialID, studentID uuid.UUID) (int, error) {
	m.materialViews[materialID] = append(m.materialViews[materialID], studentID)
	count := 0
	for _, sID := range m.materialViews[materialID] {
		if sID == studentID {
			count++
		}
	}
	return count, nil
}

func (m *mockRepository) GetViewsReport(ctx context.Context, materialID uuid.UUID) (*MaterialViewsReportResponse, error) {
	views, ok := m.materialViews[materialID]
	if !ok {
		return &MaterialViewsReportResponse{
			MaterialID:          materialID,
			TotalViews:          0,
			TotalStudentsViewed: 0,
			StudentViews:        []StudentViewRecord{},
		}, nil
	}

	distinct := make(map[uuid.UUID]int)
	for _, sID := range views {
		distinct[sID]++
	}

	records := make([]StudentViewRecord, 0, len(distinct))
	for sID, count := range distinct {
		records = append(records, StudentViewRecord{
			StudentID:    sID,
			StudentName:  "Test Student",
			StudentEmail: "student@encertia.cat",
			ViewCount:    count,
		})
	}

	return &MaterialViewsReportResponse{
		MaterialID:          materialID,
		TotalViews:          len(views),
		TotalStudentsViewed: len(distinct),
		StudentViews:        records,
	}, nil
}

func (m *mockRepository) ListUnitMaterials(ctx context.Context, unitID uuid.UUID) ([]Material, error) {
	matIDs := m.unitMaterials[unitID]
	var result []Material
	for _, id := range matIDs {
		if mat, ok := m.materials[id]; ok && mat.DeletedAt == nil {
			result = append(result, *mat)
		}
	}
	return result, nil
}

func (m *mockRepository) LinkMaterialToUnit(ctx context.Context, unitID, materialID uuid.UUID, orderIndex int) error {
	m.unitMaterials[unitID] = append(m.unitMaterials[unitID], materialID)
	key := unitID.String() + ":" + materialID.String()
	m.unitMaterialOrders[key] = orderIndex
	return nil
}

func (m *mockRepository) UnlinkMaterialFromUnit(ctx context.Context, unitID, materialID uuid.UUID) error {
	list := m.unitMaterials[unitID]
	newList := make([]uuid.UUID, 0, len(list))
	found := false
	for _, id := range list {
		if id == materialID {
			found = true
		} else {
			newList = append(newList, id)
		}
	}
	if !found {
		return ErrNotFound
	}
	m.unitMaterials[unitID] = newList
	return nil
}

func (m *mockRepository) GetCourseTeacherID(ctx context.Context, courseID uuid.UUID) (uuid.UUID, error) {
	tID, ok := m.courseTeachers[courseID]
	if !ok {
		return uuid.Nil, ErrNotFound
	}
	return tID, nil
}

func (m *mockRepository) GetCourseIDByUnitID(ctx context.Context, unitID uuid.UUID) (uuid.UUID, error) {
	cID, ok := m.unitCourses[unitID]
	if !ok {
		return uuid.Nil, ErrNotFound
	}
	return cID, nil
}
