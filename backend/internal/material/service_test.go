package material

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func setupServiceTest() (*service, *mockRepository) {
	repo := newMockRepository()
	svc := &service{repo: repo}
	return svc, repo
}

func TestService_CreateMaterial(t *testing.T) {
	svc, _ := setupServiceTest()
	ctx := context.Background()
	teacherID := uuid.New()

	t.Run("Success Document", func(t *testing.T) {
		req := CreateMaterialRequest{
			Title:        "Document Teoria 1",
			MaterialType: TypeDocument,
			FileURL:      stringPtr("http://example.com/doc.pdf"),
		}
		resp, appErr := svc.CreateMaterial(ctx, teacherID, "teacher", req)
		if appErr != nil {
			t.Fatalf("expected no error, got: %v", appErr)
		}
		if resp.Title != "Document Teoria 1" {
			t.Errorf("expected title 'Document Teoria 1', got '%s'", resp.Title)
		}
		if resp.TeacherID != teacherID {
			t.Errorf("expected teacherID %s, got %s", teacherID, resp.TeacherID)
		}
	})

	t.Run("Success Video", func(t *testing.T) {
		req := CreateMaterialRequest{
			Title:         "Video Explicatiu",
			MaterialType:  TypeVideo,
			VideoURL:      stringPtr("https://youtube.com/watch?v=12345"),
			VideoProvider: providerPtr(ProviderYouTube),
		}
		resp, appErr := svc.CreateMaterial(ctx, teacherID, "teacher", req)
		if appErr != nil {
			t.Fatalf("expected no error, got: %v", appErr)
		}
		if resp.MaterialType != TypeVideo {
			t.Errorf("expected type 'video', got '%s'", resp.MaterialType)
		}
	})

	t.Run("Validation Error Short Title", func(t *testing.T) {
		req := CreateMaterialRequest{
			Title:        "A",
			MaterialType: TypeDocument,
		}
		_, appErr := svc.CreateMaterial(ctx, teacherID, "teacher", req)
		if appErr == nil || appErr.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %v", appErr)
		}
	})

	t.Run("Validation Error Invalid Type", func(t *testing.T) {
		req := CreateMaterialRequest{
			Title:        "Material Prova",
			MaterialType: "audio",
		}
		_, appErr := svc.CreateMaterial(ctx, teacherID, "teacher", req)
		if appErr == nil || appErr.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request, got %v", appErr)
		}
	})

	t.Run("Forbidden Student", func(t *testing.T) {
		req := CreateMaterialRequest{
			Title:        "Material Prova",
			MaterialType: TypeDocument,
		}
		_, appErr := svc.CreateMaterial(ctx, uuid.New(), "student", req)
		if appErr == nil || appErr.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden, got %v", appErr)
		}
	})
}

func TestService_GetMaterialByID(t *testing.T) {
	svc, repo := setupServiceTest()
	ctx := context.Background()
	matID := uuid.New()
	teacherID := uuid.New()

	repo.materials[matID] = &Material{
		ID:           matID,
		Title:        "PDF Tema 1",
		MaterialType: TypeDocument,
		TeacherID:    teacherID,
	}

	t.Run("Success", func(t *testing.T) {
		resp, appErr := svc.GetMaterialByID(ctx, matID)
		if appErr != nil {
			t.Fatalf("expected no error, got: %v", appErr)
		}
		if resp.ID != matID {
			t.Errorf("expected id %s, got %s", matID, resp.ID)
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		_, appErr := svc.GetMaterialByID(ctx, uuid.New())
		if appErr == nil || appErr.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 Not Found, got %v", appErr)
		}
	})
}

func TestService_UpdateMaterial(t *testing.T) {
	svc, repo := setupServiceTest()
	ctx := context.Background()
	matID := uuid.New()
	teacherID := uuid.New()
	otherTeacherID := uuid.New()

	repo.materials[matID] = &Material{
		ID:           matID,
		Title:        "PDF Tema 1",
		MaterialType: TypeDocument,
		TeacherID:    teacherID,
	}

	t.Run("Success Owner", func(t *testing.T) {
		req := UpdateMaterialRequest{
			Title: stringPtr("PDF Tema 1 Actualitzat"),
		}
		resp, appErr := svc.UpdateMaterial(ctx, teacherID, "teacher", matID, req)
		if appErr != nil {
			t.Fatalf("expected no error, got: %v", appErr)
		}
		if resp.Title != "PDF Tema 1 Actualitzat" {
			t.Errorf("expected updated title, got %s", resp.Title)
		}
	})

	t.Run("Forbidden Non Owner Teacher", func(t *testing.T) {
		req := UpdateMaterialRequest{
			Title: stringPtr("Intent de modificació"),
		}
		_, appErr := svc.UpdateMaterial(ctx, otherTeacherID, "teacher", matID, req)
		if appErr == nil || appErr.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden, got %v", appErr)
		}
	})

	t.Run("Success Admin", func(t *testing.T) {
		req := UpdateMaterialRequest{
			Title: stringPtr("PDF Modificat per Admin"),
		}
		resp, appErr := svc.UpdateMaterial(ctx, uuid.New(), "admin", matID, req)
		if appErr != nil {
			t.Fatalf("expected no error, got: %v", appErr)
		}
		if resp.Title != "PDF Modificat per Admin" {
			t.Errorf("expected admin updated title, got %s", resp.Title)
		}
	})
}

func TestService_DeleteMaterial(t *testing.T) {
	svc, repo := setupServiceTest()
	ctx := context.Background()
	matID := uuid.New()
	teacherID := uuid.New()

	repo.materials[matID] = &Material{
		ID:           matID,
		Title:        "PDF per Esborrar",
		MaterialType: TypeDocument,
		TeacherID:    teacherID,
	}

	t.Run("Forbidden Non Owner", func(t *testing.T) {
		appErr := svc.DeleteMaterial(ctx, uuid.New(), "teacher", matID)
		if appErr == nil || appErr.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden, got %v", appErr)
		}
	})

	t.Run("Success Owner", func(t *testing.T) {
		appErr := svc.DeleteMaterial(ctx, teacherID, "teacher", matID)
		if appErr != nil {
			t.Fatalf("expected no error, got %v", appErr)
		}
	})
}

func TestService_RecordViewAndReport(t *testing.T) {
	svc, repo := setupServiceTest()
	ctx := context.Background()
	matID := uuid.New()
	teacherID := uuid.New()
	studentID := uuid.New()

	repo.materials[matID] = &Material{
		ID:           matID,
		Title:        "PDF Tema 2",
		MaterialType: TypeDocument,
		TeacherID:    teacherID,
	}

	t.Run("Record View", func(t *testing.T) {
		res, appErr := svc.RecordMaterialView(ctx, studentID, matID)
		if appErr != nil {
			t.Fatalf("expected no error, got %v", appErr)
		}
		if res["success"] != true {
			t.Errorf("expected success true")
		}
		if res["viewCount"] != 1 {
			t.Errorf("expected viewCount 1, got %v", res["viewCount"])
		}
	})

	t.Run("Get Report Owner Teacher", func(t *testing.T) {
		report, appErr := svc.GetMaterialViewsReport(ctx, teacherID, "teacher", matID)
		if appErr != nil {
			t.Fatalf("expected no error, got %v", appErr)
		}
		if report.TotalViews != 1 {
			t.Errorf("expected totalViews 1, got %d", report.TotalViews)
		}
	})

	t.Run("Get Report Forbidden Non Owner", func(t *testing.T) {
		_, appErr := svc.GetMaterialViewsReport(ctx, uuid.New(), "teacher", matID)
		if appErr == nil || appErr.StatusCode != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden, got %v", appErr)
		}
	})
}

func stringPtr(s string) *string {
	return &s
}

func providerPtr(p VideoProvider) *VideoProvider {
	return &p
}
