package course

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

func TestService_CreateCourse(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewService(mockRepo)
	ctx := context.Background()

	teacherID := uuid.New()
	studentID := uuid.New()

	t.Run("success by teacher", func(t *testing.T) {
		req := CreateCourseRequest{
			Title: "Matemàtiques I",
			Code:  "MATH101",
		}
		res, err := svc.CreateCourse(ctx, teacherID, "teacher", req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if res.Title != "Matemàtiques I" || res.Code != "MATH101" {
			t.Errorf("unexpected course content: %+v", res)
		}
	})

	t.Run("forbidden for student", func(t *testing.T) {
		req := CreateCourseRequest{
			Title: "Física I",
			Code:  "PHYS101",
		}
		_, err := svc.CreateCourse(ctx, studentID, "student", req)
		if err == nil || err.StatusCode != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden, got %v", err)
		}
	})

	t.Run("validation short title", func(t *testing.T) {
		req := CreateCourseRequest{
			Title: "AB",
			Code:  "MATH102",
		}
		_, err := svc.CreateCourse(ctx, teacherID, "teacher", req)
		if err == nil || err.StatusCode != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for short title, got %v", err)
		}
	})

	t.Run("duplicate course code conflict", func(t *testing.T) {
		req := CreateCourseRequest{
			Title: "Matemàtiques II",
			Code:  "MATH101",
		}
		_, err := svc.CreateCourse(ctx, teacherID, "teacher", req)
		if err == nil || err.StatusCode != http.StatusConflict {
			t.Errorf("expected 409 Conflict, got %v", err)
		}
	})
}

func TestService_CoursePermissionsAndEnrollment(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewService(mockRepo)
	ctx := context.Background()

	teacherID := uuid.New()
	otherTeacherID := uuid.New()
	studentID := uuid.New()

	cRes, _ := svc.CreateCourse(ctx, teacherID, "teacher", CreateCourseRequest{
		Title: "Química Organica",
		Code:  "CHEM201",
	})

	courseID := cRes.ID

	t.Run("other teacher cannot update course", func(t *testing.T) {
		newTitle := "Química Inorgànica"
		_, err := svc.UpdateCourse(ctx, otherTeacherID, "teacher", courseID, UpdateCourseRequest{Title: &newTitle})
		if err == nil || err.StatusCode != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden, got %v", err)
		}
	})

	t.Run("unenrolled student cannot get course detail", func(t *testing.T) {
		_, err := svc.GetCourseByID(ctx, studentID, "student", courseID)
		if err == nil || err.StatusCode != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden for unenrolled student, got %v", err)
		}
	})

	t.Run("teacher enrolls student", func(t *testing.T) {
		studentsRes, err := svc.EnrollStudents(ctx, teacherID, "teacher", courseID, EnrollStudentsRequest{
			StudentIDs: []uuid.UUID{studentID},
		})
		if err != nil {
			t.Fatalf("expected no error enrolling student, got %v", err)
		}
		if studentsRes.Total != 1 {
			t.Errorf("expected 1 enrolled student, got %d", studentsRes.Total)
		}
	})

	t.Run("enrolled student can view course detail", func(t *testing.T) {
		detail, err := svc.GetCourseByID(ctx, studentID, "student", courseID)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if detail.ID != courseID {
			t.Errorf("unexpected course ID: %v", detail.ID)
		}
	})

	t.Run("teacher unenrolls student", func(t *testing.T) {
		err := svc.UnenrollStudent(ctx, teacherID, "teacher", courseID, studentID)
		if err != nil {
			t.Fatalf("expected no error unenrolling student, got %v", err)
		}
	})
}

func TestService_CourseUnitsAndScript(t *testing.T) {
	mockRepo := newMockRepository()
	svc := NewService(mockRepo)
	ctx := context.Background()

	teacherID := uuid.New()
	cRes, _ := svc.CreateCourse(ctx, teacherID, "teacher", CreateCourseRequest{
		Title: "Història de l'Art",
		Code:  "ART101",
	})
	courseID := cRes.ID

	var unitID uuid.UUID
	t.Run("create course unit", func(t *testing.T) {
		unitRes, err := svc.CreateCourseUnit(ctx, teacherID, "teacher", courseID, CreateCourseUnitRequest{
			Title: "Tema 1: El Renaixement",
		})
		if err != nil {
			t.Fatalf("expected no error creating unit, got %v", err)
		}
		if unitRes.Title != "Tema 1: El Renaixement" {
			t.Errorf("unexpected unit title: %s", unitRes.Title)
		}
		unitID = unitRes.ID
	})

	t.Run("link quiz to unit", func(t *testing.T) {
		quizID := uuid.New()
		err := svc.LinkQuizToUnit(ctx, teacherID, "teacher", courseID, unitID, quizID)
		if err != nil {
			t.Fatalf("expected no error linking quiz, got %v", err)
		}
	})

	t.Run("update unit script", func(t *testing.T) {
		quizID := uuid.New()
		req := []CreateScriptBlockRequest{
			{
				BlockType:  "material",
				OrderIndex: 0,
				Title:      "Introducció en PDF",
			},
			{
				BlockType:  "quiz",
				OrderIndex: 1,
				Title:      "Qüestionari ràpid",
				QuizID:     &quizID,
			},
		}

		blocks, err := svc.UpdateUnitScript(ctx, teacherID, "teacher", courseID, unitID, req)
		if err != nil {
			t.Fatalf("expected no error updating script, got %v", err)
		}
		if len(blocks) != 2 {
			t.Fatalf("expected 2 script blocks, got %d", len(blocks))
		}
		if blocks[0].Title != "Introducció en PDF" {
			t.Errorf("unexpected block title: %s", blocks[0].Title)
		}
	})
}
