package course

import (
	"time"

	"github.com/google/uuid"
)

// Course represents a course entity in the system.
type Course struct {
	ID                    uuid.UUID  `json:"id"`
	Title                 string     `json:"title"`
	Code                  string     `json:"code"`
	Description           *string    `json:"description"`
	Status                string     `json:"status"` // draft, active, archived
	StartDate             *string    `json:"startDate"`
	EndDate               *string    `json:"endDate"`
	TeacherID             uuid.UUID  `json:"teacherId"`
	TeacherName           *string    `json:"teacherName"`
	EnrolledStudentsCount int        `json:"enrolledStudentsCount"`
	UnitsCount            int        `json:"unitsCount"`
	CreatedAt             time.Time  `json:"createdAt"`
	UpdatedAt             time.Time  `json:"updatedAt"`
	DeletedAt             *time.Time `json:"-"`
}

// CourseResponse matches the OpenAPI specification for course items.
type CourseResponse struct {
	ID                    uuid.UUID `json:"id"`
	Title                 string    `json:"title"`
	Code                  string    `json:"code"`
	Description           *string   `json:"description"`
	Status                string    `json:"status"`
	StartDate             *string   `json:"startDate"`
	EndDate               *string   `json:"endDate"`
	TeacherID             uuid.UUID `json:"teacherId"`
	TeacherName           *string   `json:"teacherName"`
	EnrolledStudentsCount int       `json:"enrolledStudentsCount"`
	UnitsCount            int       `json:"unitsCount"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}

// CourseListResponse represents a paginated list of courses.
type CourseListResponse struct {
	Items      []CourseResponse `json:"items"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"pageSize"`
	TotalPages int              `json:"totalPages"`
}

// CourseDetailResponse represents detailed information of a course including its units.
type CourseDetailResponse struct {
	CourseResponse
	Units []CourseUnitResponse `json:"units"`
}

// CreateCourseRequest represents payload to create a new course.
type CreateCourseRequest struct {
	Title       string  `json:"title"`
	Code        string  `json:"code"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	StartDate   *string `json:"startDate"`
	EndDate     *string `json:"endDate"`
}

// UpdateCourseRequest represents payload to update a course.
type UpdateCourseRequest struct {
	Title       *string `json:"title"`
	Code        *string `json:"code"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	StartDate   *string `json:"startDate"`
	EndDate     *string `json:"endDate"`
}

// EnrolledStudentResponse represents a student enrolled in a course.
type EnrolledStudentResponse struct {
	ID         uuid.UUID `json:"id"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	Email      string    `json:"email"`
	EnrolledAt time.Time `json:"enrolledAt"`
}

// CourseStudentsResponse represents the list of enrolled students for a course.
type CourseStudentsResponse struct {
	CourseID uuid.UUID                 `json:"courseId"`
	Total    int                       `json:"total"`
	Students []EnrolledStudentResponse `json:"students"`
}

// EnrollStudentsRequest represents payload to enroll students into a course.
type EnrollStudentsRequest struct {
	StudentIDs []uuid.UUID `json:"studentIds"`
}

// CourseUnit represents a course unit entity.
type CourseUnit struct {
	ID           uuid.UUID  `json:"id"`
	CourseID     uuid.UUID  `json:"courseId"`
	Title        string     `json:"title"`
	Description  *string    `json:"description"`
	OrderIndex   int        `json:"orderIndex"`
	QuizzesCount int        `json:"quizzesCount"`
	BlocksCount  int        `json:"blocksCount"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	DeletedAt    *time.Time `json:"-"`
}

// CourseUnitResponse matches the OpenAPI specification for unit items.
type CourseUnitResponse struct {
	ID           uuid.UUID `json:"id"`
	CourseID     uuid.UUID `json:"courseId"`
	Title        string    `json:"title"`
	Description  *string   `json:"description"`
	OrderIndex   int       `json:"orderIndex"`
	QuizzesCount int       `json:"quizzesCount"`
	BlocksCount  int       `json:"blocksCount"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// CreateCourseUnitRequest represents payload to create a course unit.
type CreateCourseUnitRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	OrderIndex  *int    `json:"orderIndex"`
}

// UpdateCourseUnitRequest represents payload to update a course unit.
type UpdateCourseUnitRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	OrderIndex  *int    `json:"orderIndex"`
}

// LinkQuizRequest represents payload to link a quiz to a course unit.
type LinkQuizRequest struct {
	QuizID uuid.UUID `json:"quizId"`
}

// LinkedQuiz represents a quiz linked to a unit.
type LinkedQuiz struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	QuestionsCount int       `json:"questionsCount"`
}

// CourseUnitDetailResponse represents unit details with linked quizzes and script blocks.
type CourseUnitDetailResponse struct {
	CourseUnitResponse
	LinkedQuizzes []LinkedQuiz          `json:"linkedQuizzes"`
	ScriptBlocks  []ScriptBlockResponse `json:"scriptBlocks"`
}

// ScriptBlock represents a class script block.
type ScriptBlock struct {
	ID              uuid.UUID  `json:"id"`
	UnitID          uuid.UUID  `json:"unitId"`
	BlockType       string     `json:"blockType"` // material, quiz, break
	OrderIndex      int        `json:"orderIndex"`
	Title           string     `json:"title"`
	Description     *string    `json:"description"`
	MaterialID      *uuid.UUID `json:"materialId"`
	PdfURL          *string    `json:"pdfUrl"`
	StartPage       *int       `json:"startPage"`
	EndPage         *int       `json:"endPage"`
	QuizID          *uuid.UUID `json:"quizId"`
	QuizTitle       *string    `json:"quizTitle"`
	DurationMinutes *int       `json:"durationMinutes"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// ScriptBlockResponse matches the OpenAPI specification for script block response.
type ScriptBlockResponse struct {
	ID              uuid.UUID  `json:"id"`
	UnitID          uuid.UUID  `json:"unitId"`
	BlockType       string     `json:"blockType"`
	OrderIndex      int        `json:"orderIndex"`
	Title           string     `json:"title"`
	Description     *string    `json:"description"`
	MaterialID      *uuid.UUID `json:"materialId"`
	PdfURL          *string    `json:"pdfUrl"`
	StartPage       *int       `json:"startPage"`
	EndPage         *int       `json:"endPage"`
	QuizID          *uuid.UUID `json:"quizId"`
	QuizTitle       *string    `json:"quizTitle"`
	DurationMinutes *int       `json:"durationMinutes"`
	CreatedAt       time.Time  `json:"createdAt"`
}

// CreateScriptBlockRequest represents payload for a single script block in sequence update.
type CreateScriptBlockRequest struct {
	BlockType       string     `json:"blockType"`
	OrderIndex      int        `json:"orderIndex"`
	Title           string     `json:"title"`
	Description     *string    `json:"description"`
	MaterialID      *uuid.UUID `json:"materialId"`
	PdfURL          *string    `json:"pdfUrl"`
	StartPage       *int       `json:"startPage"`
	EndPage         *int       `json:"endPage"`
	QuizID          *uuid.UUID `json:"quizId"`
	DurationMinutes *int       `json:"durationMinutes"`
}

// CourseListFilters holds filter parameters for listing courses.
type CourseListFilters struct {
	Page     int
	PageSize int
	Search   string
	Status   string
}
