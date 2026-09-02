package material_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/encertia/backend/internal/material"
	"github.com/encertia/backend/internal/shared"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func setupTestMaterialRouter(svc material.Service, storageSvc shared.StorageService, actorID uuid.UUID, actorRole, actorEmail string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	handler := material.NewHandler(svc, storageSvc)

	authMiddleware := func(c *gin.Context) {
		if actorID != uuid.Nil {
			c.Set(shared.CtxKeyUserID, actorID.String())
			c.Set(shared.CtxKeyUserEmail, actorEmail)
			c.Set(shared.CtxKeyUserRole, actorRole)
		}
		c.Next()
	}

	handler.RegisterRoutes(router.Group(""), authMiddleware)
	return router
}

func TestHTTP_ListAndCreateMaterials(t *testing.T) {
	svc, repo := setupMaterialService()
	storageSvc := shared.NewStorageService(shared.StorageConfig{LocalUploadDir: t.TempDir()})
	teacherID := uuid.New()
	router := setupTestMaterialRouter(svc, storageSvc, teacherID, "teacher", "teacher@encertia.cat")

	t.Run("Create Material Document", func(t *testing.T) {
		body := map[string]interface{}{
			"title":        "Apunts de Programació",
			"description":  "Tema 1",
			"materialType": "document",
			"fileUrl":      "http://localhost:8080/uploads/doc1.pdf",
			"fileName":     "doc1.pdf",
		}
		jsonBytes, _ := json.Marshal(body)

		req, _ := http.NewRequest(http.MethodPost, "/materials", bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201 Created, got %d: %s", w.Code, w.Body.String())
		}

		var res material.MaterialResponse
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to parse json response: %v", err)
		}
		if res.Title != "Apunts de Programació" {
			t.Errorf("expected title 'Apunts de Programació', got '%s'", res.Title)
		}
	})

	t.Run("List Materials", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/materials?page=1&pageSize=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}

		var res material.MaterialListResponse
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to parse json response: %v", err)
		}
		if res.Total < 1 {
			t.Errorf("expected total >= 1, got %d", res.Total)
		}
	})

	_ = repo
}

func TestHTTP_UploadMaterialFile(t *testing.T) {
	svc, _ := setupMaterialService()
	storageSvc := shared.NewStorageService(shared.StorageConfig{LocalUploadDir: t.TempDir()})
	teacherID := uuid.New()
	router := setupTestMaterialRouter(svc, storageSvc, teacherID, "teacher", "teacher@encertia.cat")

	t.Run("Upload PDF file", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "test_document.pdf")
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		_, _ = part.Write([]byte("%PDF-1.4 test pdf content /Type /Page"))
		_ = writer.Close()

		req, _ := http.NewRequest(http.MethodPost, "/materials/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}

		var res material.UploadFileResponse
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to parse json response: %v", err)
		}
		if res.FileName != "test_document.pdf" {
			t.Errorf("expected filename test_document.pdf, got %s", res.FileName)
		}
	})
}

func TestHTTP_GetUpdateDeleteMaterial(t *testing.T) {
	svc, _ := setupMaterialService()
	storageSvc := shared.NewStorageService(shared.StorageConfig{LocalUploadDir: t.TempDir()})
	teacherID := uuid.New()
	router := setupTestMaterialRouter(svc, storageSvc, teacherID, "teacher", "teacher@encertia.cat")

	ctx := t.Context()
	res, _ := svc.CreateMaterial(ctx, teacherID, "teacher", material.CreateMaterialRequest{
		Title:        "Material Inicial",
		MaterialType: material.TypeDocument,
	})

	matID := res.ID

	t.Run("Get Material", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/materials/"+matID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Update Material", func(t *testing.T) {
		body := map[string]interface{}{
			"title": "Material Actualitzat",
		}
		jsonBytes, _ := json.Marshal(body)

		req, _ := http.NewRequest(http.MethodPut, "/materials/"+matID.String(), bytes.NewBuffer(jsonBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Delete Material", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodDelete, "/materials/"+matID.String(), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204 No Content, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestHTTP_MaterialViewsAndReports(t *testing.T) {
	svc, _ := setupMaterialService()
	storageSvc := shared.NewStorageService(shared.StorageConfig{LocalUploadDir: t.TempDir()})
	teacherID := uuid.New()
	studentID := uuid.New()

	ctx := t.Context()
	res, _ := svc.CreateMaterial(ctx, teacherID, "teacher", material.CreateMaterialRequest{
		Title:        "Material per a Visualitzar",
		MaterialType: material.TypeDocument,
	})
	matID := res.ID

	t.Run("Record View by Student", func(t *testing.T) {
		studentRouter := setupTestMaterialRouter(svc, storageSvc, studentID, "student", "student@encertia.cat")
		req, _ := http.NewRequest(http.MethodPost, "/materials/"+matID.String()+"/views", nil)
		w := httptest.NewRecorder()
		studentRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Get Views Report by Teacher", func(t *testing.T) {
		teacherRouter := setupTestMaterialRouter(svc, storageSvc, teacherID, "teacher", "teacher@encertia.cat")
		req, _ := http.NewRequest(http.MethodGet, "/materials/"+matID.String()+"/views", nil)
		w := httptest.NewRecorder()
		teacherRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func setupMaterialService() (material.Service, *mockRepository) {
	repo := newMockRepository()
	svc := material.NewService(repo)
	return svc, repo
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		materials:          make(map[uuid.UUID]*material.Material),
		unitMaterials:      make(map[uuid.UUID][]uuid.UUID),
		unitMaterialOrders: make(map[string]int),
		materialViews:      make(map[uuid.UUID][]uuid.UUID),
		courseTeachers:     make(map[uuid.UUID]uuid.UUID),
		unitCourses:        make(map[uuid.UUID]uuid.UUID),
	}
}

type mockRepository struct {
	materials          map[uuid.UUID]*material.Material
	unitMaterials      map[uuid.UUID][]uuid.UUID
	unitMaterialOrders map[string]int
	materialViews      map[uuid.UUID][]uuid.UUID
	courseTeachers     map[uuid.UUID]uuid.UUID
	unitCourses        map[uuid.UUID]uuid.UUID
}

func (m *mockRepository) Create(ctx context.Context, mat *material.Material) error {
	if mat.ID == uuid.Nil {
		mat.ID = uuid.New()
	}
	m.materials[mat.ID] = mat
	return nil
}

func (m *mockRepository) GetByID(ctx context.Context, id uuid.UUID) (*material.Material, error) {
	mat, ok := m.materials[id]
	if !ok || mat.DeletedAt != nil {
		return nil, material.ErrNotFound
	}
	return mat, nil
}

func (m *mockRepository) Update(ctx context.Context, mat *material.Material) error {
	existing, ok := m.materials[mat.ID]
	if !ok || existing.DeletedAt != nil {
		return material.ErrNotFound
	}
	m.materials[mat.ID] = mat
	return nil
}

func (m *mockRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	mat, ok := m.materials[id]
	if !ok || mat.DeletedAt != nil {
		return material.ErrNotFound
	}
	mat.DeletedAt = &mat.CreatedAt
	return nil
}

func (m *mockRepository) List(ctx context.Context, filters material.MaterialListFilters, teacherID *uuid.UUID) ([]material.Material, int, error) {
	var result []material.Material
	for _, mat := range m.materials {
		if mat.DeletedAt != nil {
			continue
		}
		if teacherID != nil && mat.TeacherID != *teacherID {
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

func (m *mockRepository) GetViewsReport(ctx context.Context, materialID uuid.UUID) (*material.MaterialViewsReportResponse, error) {
	views := m.materialViews[materialID]
	return &material.MaterialViewsReportResponse{
		MaterialID:          materialID,
		TotalViews:          len(views),
		TotalStudentsViewed: len(views),
		StudentViews:        []material.StudentViewRecord{},
	}, nil
}

func (m *mockRepository) ListUnitMaterials(ctx context.Context, unitID uuid.UUID) ([]material.Material, error) {
	matIDs := m.unitMaterials[unitID]
	var result []material.Material
	for _, id := range matIDs {
		if mat, ok := m.materials[id]; ok && mat.DeletedAt == nil {
			result = append(result, *mat)
		}
	}
	return result, nil
}

func (m *mockRepository) LinkMaterialToUnit(ctx context.Context, unitID, materialID uuid.UUID, orderIndex int) error {
	m.unitMaterials[unitID] = append(m.unitMaterials[unitID], materialID)
	return nil
}

func (m *mockRepository) UnlinkMaterialFromUnit(ctx context.Context, unitID, materialID uuid.UUID) error {
	return nil
}

func (m *mockRepository) GetCourseTeacherID(ctx context.Context, courseID uuid.UUID) (uuid.UUID, error) {
	return m.courseTeachers[courseID], nil
}

func (m *mockRepository) GetCourseIDByUnitID(ctx context.Context, unitID uuid.UUID) (uuid.UUID, error) {
	return m.unitCourses[unitID], nil
}
