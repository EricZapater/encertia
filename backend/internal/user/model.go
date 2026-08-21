package user

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user in Encertia.
type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleTeacher UserRole = "teacher"
	RoleStudent UserRole = "student"
)

// IsValid checks if the UserRole is valid according to the OpenAPI enum.
func (r UserRole) IsValid() bool {
	return r == RoleAdmin || r == RoleTeacher || r == RoleStudent
}

// User represents the public user entity as defined in OpenAPI.
type User struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	Role      UserRole   `json:"role"`
	IsActive  bool       `json:"isActive"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

// UserDB represents the user entity stored in PostgreSQL.
type UserDB struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	FirstName    string     `json:"firstName"`
	LastName     string     `json:"lastName"`
	Role         UserRole   `json:"role"`
	IsActive     bool       `json:"isActive"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	DeletedAt    *time.Time `json:"-"`
}

// ToUser converts UserDB to public User model.
func (u *UserDB) ToUser() User {
	var updatedAt *time.Time
	if !u.UpdatedAt.IsZero() {
		updatedAt = &u.UpdatedAt
	}
	return User{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: updatedAt,
	}
}

// CreateUserInput defines payload for POST /users.
type CreateUserInput struct {
	Email     string   `json:"email"`
	Password  string   `json:"password"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Role      UserRole `json:"role"`
}

// UpdateUserInput defines payload for PUT /users/{id}.
type UpdateUserInput struct {
	Email     *string   `json:"email,omitempty"`
	FirstName *string   `json:"firstName,omitempty"`
	LastName  *string   `json:"lastName,omitempty"`
	Role      *UserRole `json:"role,omitempty"`
	IsActive  *bool     `json:"isActive,omitempty"`
}

// ResetPasswordInput defines payload for POST /users/{id}/password.
type ResetPasswordInput struct {
	NewPassword string `json:"newPassword"`
}

// BatchUserItem defines a single item in a batch user creation request.
type BatchUserItem struct {
	Email     string   `json:"email"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Role      UserRole `json:"role"`
	Password  *string  `json:"password,omitempty"`
}

// BatchCreateUsersRequest defines payload for POST /users/batch.
type BatchCreateUsersRequest struct {
	Users []BatchUserItem `json:"users"`
}

// BatchItemError represents an error on a specific row during batch processing.
type BatchItemError struct {
	Row   int    `json:"row"`
	Email string `json:"email"`
	Error string `json:"error"`
}

// BatchCreateUsersResponse defines the response payload for batch import.
type BatchCreateUsersResponse struct {
	TotalRequested int              `json:"totalRequested"`
	CreatedCount   int              `json:"createdCount"`
	FailedCount    int              `json:"failedCount"`
	CreatedUsers   []User           `json:"createdUsers"`
	Errors         []BatchItemError `json:"errors"`
}

// PaginationMetadata defines pagination information.
type PaginationMetadata struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	TotalCount int `json:"totalCount"`
	TotalPages int `json:"totalPages"`
}

// UserListResponse defines the payload for GET /users.
type UserListResponse struct {
	Items      []User             `json:"items"`
	Pagination PaginationMetadata `json:"pagination"`
}

// UserResponse defines the payload for endpoints returning a single user.
type UserResponse struct {
	User User `json:"user"`
}

// ListUsersFilter defines query filters for listing users.
type ListUsersFilter struct {
	Page      int
	PageSize  int
	Search    string
	Role      string
	Status    string
	ActorRole string
}
