package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	Language  string     `json:"language"`
	IsActive  bool       `json:"isActive"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

// UserDB represents the user entity stored in the database.
type UserDB struct {
	ID           uuid.UUID  `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	FirstName    string     `json:"firstName"`
	LastName     string     `json:"lastName"`
	Role         UserRole   `json:"role"`
	Language     string     `json:"language"`
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
		Language:  u.Language,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: updatedAt,
	}
}

// RefreshTokenDB represents a refresh token persisted in PostgreSQL.
type RefreshTokenDB struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"userId"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expiresAt"`
	RevokedAt *time.Time `json:"revokedAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"-"`
}

// RegisterRequest defines the input payload for POST /auth/register.
type RegisterRequest struct {
	Email     string   `json:"email"`
	Password  string   `json:"password"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Role      UserRole `json:"role"`
	Language  string   `json:"language"`
}

// LoginRequest defines the input payload for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshTokenRequest defines the input payload for POST /auth/refresh.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// LogoutRequest defines the input payload for POST /auth/logout.
type LogoutRequest struct {
	RefreshToken *string `json:"refreshToken,omitempty"`
}

// TokenPair represents the access and refresh token pair.
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int    `json:"expiresIn"`
}

// AuthResponse represents the response for register and login.
type AuthResponse struct {
	User   User      `json:"user"`
	Tokens TokenPair `json:"tokens"`
}

// UserResponse represents the response for GET /auth/me.
type UserResponse struct {
	User User `json:"user"`
}

// MessageResponse represents generic success messages (e.g. logout).
type MessageResponse struct {
	Message string `json:"message"`
}

// JWTClaims represents custom claims inside the Encertia access token.
type JWTClaims struct {
	UserID string   `json:"userId"`
	Email  string   `json:"email"`
	Role   UserRole `json:"role"`
	jwt.RegisteredClaims
}
