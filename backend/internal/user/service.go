package user

import (
	"context"
	"errors"
	"math"
	"net/mail"
	"strings"
	"time"

	"github.com/encertia/backend/internal/shared"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Service defines business logic for the User domain.
type Service interface {
	ListUsers(ctx context.Context, actorID uuid.UUID, actorRole string, filter ListUsersFilter) (*UserListResponse, *shared.AppError)
	CreateUser(ctx context.Context, actorRole string, input CreateUserInput) (*UserResponse, *shared.AppError)
	BatchCreateUsers(ctx context.Context, actorRole string, req BatchCreateUsersRequest) (*BatchCreateUsersResponse, *shared.AppError)
	GetUserByID(ctx context.Context, actorID uuid.UUID, actorRole string, targetID uuid.UUID) (*UserResponse, *shared.AppError)
	UpdateUser(ctx context.Context, actorID uuid.UUID, actorRole string, targetID uuid.UUID, input UpdateUserInput) (*UserResponse, *shared.AppError)
	ResetPassword(ctx context.Context, actorID uuid.UUID, actorRole string, targetID uuid.UUID, input ResetPasswordInput) (*shared.MessageResponse, *shared.AppError)
	DeleteUser(ctx context.Context, actorID uuid.UUID, actorRole string, targetID uuid.UUID) (*shared.MessageResponse, *shared.AppError)
}

type userService struct {
	repo Repository
}

// NewService creates a new instance of User service.
func NewService(repo Repository) Service {
	return &userService{repo: repo}
}

func (s *userService) ListUsers(ctx context.Context, actorID uuid.UUID, actorRole string, filter ListUsersFilter) (*UserListResponse, *shared.AppError) {
	// RBAC validation
	if actorRole == string(RoleStudent) {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permisos suficients per llistar usuaris.")
	}
	if actorRole != string(RoleAdmin) && actorRole != string(RoleTeacher) {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "Rol no autoritzat.")
	}

	filter.ActorRole = actorRole

	// Pagination defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	if filter.PageSize > 100 {
		filter.PageSize = 100
	}
	if filter.Status == "" {
		filter.Status = "active"
	}

	usersDB, totalCount, err := s.repo.ListUsers(ctx, filter)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	items := make([]User, len(usersDB))
	for i, u := range usersDB {
		items[i] = u.ToUser()
	}

	totalPages := 0
	if totalCount > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(filter.PageSize)))
	}

	return &UserListResponse{
		Items: items,
		Pagination: PaginationMetadata{
			Page:       filter.Page,
			PageSize:   filter.PageSize,
			TotalCount: totalCount,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *userService) CreateUser(ctx context.Context, actorRole string, input CreateUserInput) (*UserResponse, *shared.AppError) {
	// 1. Sanitize & Validate inputs
	input.Email = strings.TrimSpace(strings.ToLower(input.Email))
	input.FirstName = strings.TrimSpace(input.FirstName)
	input.LastName = strings.TrimSpace(input.LastName)

	if input.Email == "" {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El correu electrònic és obligatori.", map[string]interface{}{"field": "email"})
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El format del correu electrònic no és vàlid.", map[string]interface{}{"field": "email"})
	}
	if len(input.Password) < 8 {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "La contrasenya ha de tenir com a mínim 8 caràcters.", map[string]interface{}{"field": "password"})
	}
	if input.FirstName == "" {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El nom és obligatori.", map[string]interface{}{"field": "firstName"})
	}
	if input.LastName == "" {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "Els cognoms són obligatoris.", map[string]interface{}{"field": "lastName"})
	}
	if !input.Role.IsValid() {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El rol especificat no és vàlid.", map[string]interface{}{"field": "role"})
	}

	// 2. RBAC check
	if actorRole == string(RoleTeacher) {
		if input.Role != RoleStudent {
			return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "Els professors només poden crear usuaris amb rol alumne.")
		}
	} else if actorRole != string(RoleAdmin) {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permisos suficients per crear usuaris.")
	}

	// 3. Check email uniqueness
	existing, err := s.repo.GetUserByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, shared.ErrInternal(err)
	}
	if existing != nil {
		return nil, shared.ErrConflict(shared.ErrCodeEmailAlreadyExists, "L'adreça de correu ja existeix.")
	}

	// 4. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	// 5. Insert user
	userDB := &UserDB{
		ID:           uuid.New(),
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Role:         input.Role,
		IsActive:     true,
	}

	if err := s.repo.CreateUser(ctx, userDB); err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, shared.ErrConflict(shared.ErrCodeEmailAlreadyExists, "L'adreça de correu ja existeix.")
		}
		return nil, shared.ErrInternal(err)
	}

	return &UserResponse{
		User: userDB.ToUser(),
	}, nil
}

func (s *userService) BatchCreateUsers(ctx context.Context, actorRole string, req BatchCreateUsersRequest) (*BatchCreateUsersResponse, *shared.AppError) {
	// RBAC check
	if actorRole != string(RoleAdmin) && actorRole != string(RoleTeacher) {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permisos suficients per realitzar importació massiva.")
	}

	if len(req.Users) == 0 {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El lot d'usuaris no pot estar buit.", nil)
	}
	if len(req.Users) > 500 {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El lot d'usuaris no pot superar els 500 registres.", nil)
	}

	createdUsers := make([]User, 0, len(req.Users))
	errorsList := make([]BatchItemError, 0)
	seenEmailsInBatch := make(map[string]bool)

	for i, item := range req.Users {
		rowNum := i + 1
		email := strings.TrimSpace(strings.ToLower(item.Email))
		firstName := strings.TrimSpace(item.FirstName)
		lastName := strings.TrimSpace(item.LastName)
		role := item.Role
		if role == "" {
			role = RoleStudent
		}

		if email == "" {
			errorsList = append(errorsList, BatchItemError{
				Row:   rowNum,
				Email: item.Email,
				Error: "EMAIL_REQUIRED",
			})
			continue
		}

		if _, err := mail.ParseAddress(email); err != nil {
			errorsList = append(errorsList, BatchItemError{
				Row:   rowNum,
				Email: email,
				Error: "INVALID_EMAIL_FORMAT",
			})
			continue
		}

		if seenEmailsInBatch[email] {
			errorsList = append(errorsList, BatchItemError{
				Row:   rowNum,
				Email: email,
				Error: "EMAIL_DUPLICATE_IN_BATCH",
			})
			continue
		}

		if firstName == "" || lastName == "" {
			errorsList = append(errorsList, BatchItemError{
				Row:   rowNum,
				Email: email,
				Error: "MISSING_NAME_FIELDS",
			})
			continue
		}

		if !role.IsValid() {
			errorsList = append(errorsList, BatchItemError{
				Row:   rowNum,
				Email: email,
				Error: "INVALID_ROLE",
			})
			continue
		}

		if actorRole == string(RoleTeacher) && role != RoleStudent {
			errorsList = append(errorsList, BatchItemError{
				Row:   rowNum,
				Email: email,
				Error: "TEACHER_CAN_ONLY_CREATE_STUDENTS",
			})
			continue
		}

		pwd := "Provisional123!"
		if item.Password != nil && strings.TrimSpace(*item.Password) != "" {
			if len(*item.Password) < 8 {
				errorsList = append(errorsList, BatchItemError{
					Row:   rowNum,
					Email: email,
					Error: "PASSWORD_TOO_SHORT",
				})
				continue
			}
			pwd = *item.Password
		}

		// Check if email already exists in DB
		existing, err := s.repo.GetUserByEmail(ctx, email)
		if err != nil && !errors.Is(err, ErrUserNotFound) {
			errorsList = append(errorsList, BatchItemError{
				Row:   rowNum,
				Email: email,
				Error: "DATABASE_ERROR",
			})
			continue
		}
		if existing != nil {
			errorsList = append(errorsList, BatchItemError{
				Row:   rowNum,
				Email: email,
				Error: "EMAIL_ALREADY_EXISTS",
			})
			continue
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err != nil {
			errorsList = append(errorsList, BatchItemError{
				Row:   rowNum,
				Email: email,
				Error: "PASSWORD_HASH_ERROR",
			})
			continue
		}

		userDB := &UserDB{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: string(hashedPassword),
			FirstName:    firstName,
			LastName:     lastName,
			Role:         role,
			IsActive:     true,
		}

		if err := s.repo.CreateUser(ctx, userDB); err != nil {
			errorsList = append(errorsList, BatchItemError{
				Row:   rowNum,
				Email: email,
				Error: "CREATION_FAILED",
			})
			continue
		}

		seenEmailsInBatch[email] = true
		createdUsers = append(createdUsers, userDB.ToUser())
	}

	return &BatchCreateUsersResponse{
		TotalRequested: len(req.Users),
		CreatedCount:   len(createdUsers),
		FailedCount:    len(errorsList),
		CreatedUsers:   createdUsers,
		Errors:         errorsList,
	}, nil
}

func (s *userService) GetUserByID(ctx context.Context, actorID uuid.UUID, actorRole string, targetID uuid.UUID) (*UserResponse, *shared.AppError) {
	userDB, err := s.repo.GetUserByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, shared.ErrNotFound(shared.ErrCodeUserNotFound, "Usuari no trobat o donat de baixa.")
		}
		return nil, shared.ErrInternal(err)
	}

	// RBAC validation
	if actorRole == string(RoleAdmin) {
		// Admin can view any user
	} else if actorRole == string(RoleTeacher) {
		// Teacher can view students or their own profile
		if userDB.Role != RoleStudent && actorID != targetID {
			return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permisos per consultar aquest usuari.")
		}
	} else {
		// Student can only view their own profile
		if actorID != targetID {
			return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permisos per consultar aquest usuari.")
		}
	}

	return &UserResponse{
		User: userDB.ToUser(),
	}, nil
}

func (s *userService) UpdateUser(ctx context.Context, actorID uuid.UUID, actorRole string, targetID uuid.UUID, input UpdateUserInput) (*UserResponse, *shared.AppError) {
	userDB, err := s.repo.GetUserByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, shared.ErrNotFound(shared.ErrCodeUserNotFound, "Usuari no trobat.")
		}
		return nil, shared.ErrInternal(err)
	}

	// RBAC & Account status checks
	if actorRole != string(RoleAdmin) {
		// Non-admin can only update their own profile
		if actorID != targetID {
			return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permisos per modificar aquest usuari.")
		}
		// Inactive user cannot edit own profile
		if !userDB.IsActive {
			return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "El compte està inactiu i no pot modificar les seves dades.")
		}
		// Non-admin cannot modify role or isActive
		if input.Role != nil || input.IsActive != nil {
			return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "Només un administrador pot modificar el rol o l'estat d'activació.")
		}
	}

	// Apply email change
	if input.Email != nil {
		newEmail := strings.TrimSpace(strings.ToLower(*input.Email))
		if newEmail == "" {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El correu electrònic no pot estar buit.", map[string]interface{}{"field": "email"})
		}
		if _, err := mail.ParseAddress(newEmail); err != nil {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El format del correu electrònic no és vàlid.", map[string]interface{}{"field": "email"})
		}
		if newEmail != userDB.Email {
			existing, err := s.repo.GetUserByEmail(ctx, newEmail)
			if err != nil && !errors.Is(err, ErrUserNotFound) {
				return nil, shared.ErrInternal(err)
			}
			if existing != nil && existing.ID != userDB.ID {
				return nil, shared.ErrConflict(shared.ErrCodeEmailAlreadyExists, "El correu electrònic ja està en ús per un altre usuari.")
			}
			userDB.Email = newEmail
		}
	}

	// Apply name changes
	if input.FirstName != nil {
		fn := strings.TrimSpace(*input.FirstName)
		if fn == "" {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El nom no pot estar buit.", map[string]interface{}{"field": "firstName"})
		}
		userDB.FirstName = fn
	}
	if input.LastName != nil {
		ln := strings.TrimSpace(*input.LastName)
		if ln == "" {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "Els cognoms no poden estar buits.", map[string]interface{}{"field": "lastName"})
		}
		userDB.LastName = ln
	}

	// Apply admin-only fields
	if input.Role != nil {
		if !input.Role.IsValid() {
			return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "Rol no vàlid.", map[string]interface{}{"field": "role"})
		}
		userDB.Role = *input.Role
	}
	if input.IsActive != nil {
		userDB.IsActive = *input.IsActive
	}

	if err := s.repo.UpdateUser(ctx, userDB); err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, shared.ErrConflict(shared.ErrCodeEmailAlreadyExists, "El correu electrònic ja està en ús per un altre usuari.")
		}
		return nil, shared.ErrInternal(err)
	}

	// If account is deactivated, revoke sessions
	if !userDB.IsActive {
		_ = s.repo.RevokeAllUserTokens(ctx, userDB.ID, time.Now().UTC())
	}

	return &UserResponse{
		User: userDB.ToUser(),
	}, nil
}

func (s *userService) ResetPassword(ctx context.Context, actorID uuid.UUID, actorRole string, targetID uuid.UUID, input ResetPasswordInput) (*shared.MessageResponse, *shared.AppError) {
	if len(input.NewPassword) < 8 {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "La contrasenya ha de tenir com a mínim 8 caràcters.", map[string]interface{}{"field": "newPassword"})
	}

	userDB, err := s.repo.GetUserByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, shared.ErrNotFound(shared.ErrCodeUserNotFound, "Usuari no trobat.")
		}
		return nil, shared.ErrInternal(err)
	}

	// RBAC check
	if actorRole == string(RoleAdmin) {
		// Admin can reset anyone's password
	} else if actorRole == string(RoleTeacher) {
		// Teacher can only reset students' passwords
		if userDB.Role != RoleStudent {
			return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "Els professors només poden restablir la contrasenya d'alumnes.")
		}
	} else {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "No tens permisos suficients per restablir contrasenyes.")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	if err := s.repo.UpdatePassword(ctx, targetID, string(hashedPassword)); err != nil {
		return nil, shared.ErrInternal(err)
	}

	// Invalidate active sessions
	_ = s.repo.RevokeAllUserTokens(ctx, targetID, time.Now().UTC())

	return &shared.MessageResponse{
		Message: "Contrasenya actualitzada correctament.",
	}, nil
}

func (s *userService) DeleteUser(ctx context.Context, actorID uuid.UUID, actorRole string, targetID uuid.UUID) (*shared.MessageResponse, *shared.AppError) {
	// Only Admin can delete users
	if actorRole != string(RoleAdmin) {
		return nil, shared.ErrForbidden(shared.ErrCodeForbidden, "Només els administradors poden donar de baixa usuaris.")
	}

	_, err := s.repo.GetUserByID(ctx, targetID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, shared.ErrNotFound(shared.ErrCodeUserNotFound, "Usuari no trobat.")
		}
		return nil, shared.ErrInternal(err)
	}

	now := time.Now().UTC()
	if err := s.repo.SoftDeleteUser(ctx, targetID, now); err != nil {
		return nil, shared.ErrInternal(err)
	}

	// Invalidate all tokens
	_ = s.repo.RevokeAllUserTokens(ctx, targetID, now)

	return &shared.MessageResponse{
		Message: "Usuari donat de baixa correctament.",
	}, nil
}
