package group

import (
	"time"

	"github.com/google/uuid"
)

// Group represents a student group domain model.
type Group struct {
	ID           uuid.UUID  `json:"id"`
	Name         string     `json:"name"`
	AcademicYear string     `json:"academicYear"`
	CourseID     *uuid.UUID `json:"courseId"`
	CourseTitle  *string    `json:"courseTitle"`
	TeacherID    uuid.UUID  `json:"teacherId"`
	StudentCount int        `json:"studentCount"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	DeletedAt    *time.Time `json:"-"`
}

// GroupResponse matches OpenAPI response schema for single group detail.
type GroupResponse struct {
	Data Group `json:"data"`
}

// GroupListResponse matches OpenAPI response schema for paginated group list.
type GroupListResponse struct {
	Items      []Group `json:"items"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	PageSize   int     `json:"pageSize"`
	TotalPages int     `json:"totalPages"`
}

// CreateGroupRequest matches OpenAPI request schema to create a group.
type CreateGroupRequest struct {
	Name         string     `json:"name"`
	AcademicYear *string    `json:"academicYear"`
	CourseID     *uuid.UUID `json:"courseId"`
}

// UpdateGroupRequest matches OpenAPI request schema to update a group.
type UpdateGroupRequest struct {
	Name         *string    `json:"name"`
	AcademicYear *string    `json:"academicYear"`
	CourseID     *uuid.UUID `json:"courseId"`
}

// AddStudentsToGroupRequest matches OpenAPI request schema to add students to a group.
type AddStudentsToGroupRequest struct {
	StudentIDs []uuid.UUID `json:"studentIds"`
}

// GroupStudent matches OpenAPI schema for student assigned to a group.
type GroupStudent struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	FullName   string    `json:"fullName"`
	AssignedAt time.Time `json:"assignedAt"`
}

// GroupStudentListResponse matches OpenAPI schema for list of group students.
type GroupStudentListResponse struct {
	Items []GroupStudent `json:"items"`
	Total int            `json:"total"`
}

// GroupListFilters represents filter parameters for querying groups.
type GroupListFilters struct {
	Page         int
	PageSize     int
	Search       string
	AcademicYear string
	CourseID     *uuid.UUID
}
