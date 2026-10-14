package auth_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/encertia/backend/internal/auth"
	"github.com/encertia/backend/internal/shared"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func setupService() (auth.Service, *mockRepository) {
	repo := newMockRepository()
	cfg := auth.Config{
		JWTSecret:            "test-secret-key-32-characters-long!",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "encertia-test",
	}
	svc := auth.NewService(repo, cfg)
	return svc, repo
}

func TestRegister_Success(t *testing.T) {
	svc, _ := setupService()
	ctx := context.Background()

	req := auth.RegisterRequest{
		Email:     "maria.garcia@encertia.cat",
		Password:  "SecretPassword123!",
		FirstName: "Maria",
		LastName:  "Garcia",
		Role:      auth.RoleStudent,
	}

	res, appErr := svc.Register(ctx, req)
	if appErr != nil {
		t.Fatalf("unexpected error registering user: %v", appErr)
	}

	if res.User.Email != "maria.garcia@encertia.cat" {
		t.Errorf("expected email maria.garcia@encertia.cat, got %s", res.User.Email)
	}
	if res.User.FirstName != "Maria" {
		t.Errorf("expected firstName Maria, got %s", res.User.FirstName)
	}
	if res.User.Role != auth.RoleStudent {
		t.Errorf("expected role student, got %s", res.User.Role)
	}
	if !res.User.IsActive {
		t.Errorf("expected isActive true, got %v", res.User.IsActive)
	}
	if res.Tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if res.Tokens.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if res.Tokens.TokenType != "Bearer" {
		t.Errorf("expected tokenType Bearer, got %s", res.Tokens.TokenType)
	}
	if res.Tokens.ExpiresIn != 900 {
		t.Errorf("expected expiresIn 900, got %d", res.Tokens.ExpiresIn)
	}
}

func TestRegister_ForcesStudentRole(t *testing.T) {
	svc, _ := setupService()
	ctx := context.Background()

	req := auth.RegisterRequest{
		Email:     "forced.student@encertia.cat",
		Password:  "SecretPassword123!",
		FirstName: "Pol",
		LastName:  "Vila",
		Role:      auth.RoleTeacher, // Even if teacher is passed in public registration
	}

	res, appErr := svc.Register(ctx, req)
	if appErr != nil {
		t.Fatalf("unexpected error registering user: %v", appErr)
	}

	if res.User.Role != auth.RoleStudent {
		t.Errorf("expected role student for public registration, got %s", res.User.Role)
	}
}

func TestRegister_ValidationErrors(t *testing.T) {
	svc, _ := setupService()
	ctx := context.Background()

	tests := []struct {
		name string
		req  auth.RegisterRequest
		code string
	}{
		{
			name: "empty email",
			req: auth.RegisterRequest{
				Email:     "",
				Password:  "password123",
				FirstName: "Nom",
				LastName:  "Cognom",
				Role:      auth.RoleStudent,
			},
			code: shared.ErrCodeValidation,
		},
		{
			name: "invalid email format",
			req: auth.RegisterRequest{
				Email:     "invalid-email-address",
				Password:  "password123",
				FirstName: "Nom",
				LastName:  "Cognom",
				Role:      auth.RoleStudent,
			},
			code: shared.ErrCodeValidation,
		},
		{
			name: "short password",
			req: auth.RegisterRequest{
				Email:     "test@encertia.cat",
				Password:  "short",
				FirstName: "Nom",
				LastName:  "Cognom",
				Role:      auth.RoleStudent,
			},
			code: shared.ErrCodeValidation,
		},
		{
			name: "empty first name",
			req: auth.RegisterRequest{
				Email:     "test@encertia.cat",
				Password:  "password123",
				FirstName: "",
				LastName:  "Cognom",
				Role:      auth.RoleStudent,
			},
			code: shared.ErrCodeValidation,
		},
		{
			name: "empty last name",
			req: auth.RegisterRequest{
				Email:     "test@encertia.cat",
				Password:  "password123",
				FirstName: "Nom",
				LastName:  "",
				Role:      auth.RoleStudent,
			},
			code: shared.ErrCodeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, appErr := svc.Register(ctx, tt.req)
			if appErr == nil {
				t.Fatalf("expected error for %s, got nil", tt.name)
			}
			if appErr.Code != tt.code {
				t.Errorf("expected code %s, got %s", tt.code, appErr.Code)
			}
		})
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, _ := setupService()
	ctx := context.Background()

	req := auth.RegisterRequest{
		Email:     "duplicate@encertia.cat",
		Password:  "Secret123!",
		FirstName: "Joan",
		LastName:  "Puig",
		Role:      auth.RoleStudent,
	}

	_, appErr := svc.Register(ctx, req)
	if appErr != nil {
		t.Fatalf("unexpected error in first registration: %v", appErr)
	}

	_, appErr2 := svc.Register(ctx, req)
	if appErr2 == nil {
		t.Fatal("expected conflict error on duplicate email, got nil")
	}
	if appErr2.Code != shared.ErrCodeEmailAlreadyExists {
		t.Errorf("expected code EMAIL_ALREADY_EXISTS, got %s", appErr2.Code)
	}
	if appErr2.StatusCode != 409 {
		t.Errorf("expected status code 409, got %d", appErr2.StatusCode)
	}
}

func TestLogin_Success(t *testing.T) {
	svc, _ := setupService()
	ctx := context.Background()

	_, _ = svc.Register(ctx, auth.RegisterRequest{
		Email:     "login@encertia.cat",
		Password:  "CorrectPassword123!",
		FirstName: "Anna",
		LastName:  "Marti",
		Role:      auth.RoleStudent,
	})

	res, appErr := svc.Login(ctx, auth.LoginRequest{
		Email:    "login@encertia.cat",
		Password: "CorrectPassword123!",
	})
	if appErr != nil {
		t.Fatalf("unexpected login error: %v", appErr)
	}

	if res.User.Email != "login@encertia.cat" {
		t.Errorf("expected email login@encertia.cat, got %s", res.User.Email)
	}
	if res.Tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	svc, _ := setupService()
	ctx := context.Background()

	_, _ = svc.Register(ctx, auth.RegisterRequest{
		Email:     "user@encertia.cat",
		Password:  "CorrectPassword123!",
		FirstName: "Anna",
		LastName:  "Marti",
		Role:      auth.RoleStudent,
	})

	// Wrong password
	_, appErr := svc.Login(ctx, auth.LoginRequest{
		Email:    "user@encertia.cat",
		Password: "WrongPassword!",
	})
	if appErr == nil {
		t.Fatal("expected unauthorized error for wrong password, got nil")
	}
	if appErr.Code != shared.ErrCodeInvalidCredentials {
		t.Errorf("expected INVALID_CREDENTIALS, got %s", appErr.Code)
	}
	if appErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", appErr.StatusCode)
	}

	// Non-existent user
	_, appErr2 := svc.Login(ctx, auth.LoginRequest{
		Email:    "nonexistent@encertia.cat",
		Password: "AnyPassword123!",
	})
	if appErr2 == nil {
		t.Fatal("expected unauthorized error for nonexistent user, got nil")
	}
	if appErr2.Code != shared.ErrCodeInvalidCredentials {
		t.Errorf("expected INVALID_CREDENTIALS, got %s", appErr2.Code)
	}
}

func TestLogin_InactiveUser(t *testing.T) {
	svc, repo := setupService()
	ctx := context.Background()

	regRes, _ := svc.Register(ctx, auth.RegisterRequest{
		Email:     "inactive@encertia.cat",
		Password:  "CorrectPassword123!",
		FirstName: "Pau",
		LastName:  "Vidal",
		Role:      auth.RoleStudent,
	})

	// Deactivate user in repository
	userDB, _ := repo.GetUserByID(ctx, regRes.User.ID)
	userDB.IsActive = false

	_, appErr := svc.Login(ctx, auth.LoginRequest{
		Email:    "inactive@encertia.cat",
		Password: "CorrectPassword123!",
	})
	if appErr == nil {
		t.Fatal("expected unauthorized error for inactive user, got nil")
	}
	if appErr.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", appErr.StatusCode)
	}
}

func TestRefreshToken_SuccessAndRotation(t *testing.T) {
	svc, _ := setupService()
	ctx := context.Background()

	regRes, _ := svc.Register(ctx, auth.RegisterRequest{
		Email:     "refresh@encertia.cat",
		Password:  "Password123!",
		FirstName: "Pau",
		LastName:  "Roca",
		Role:      auth.RoleTeacher,
	})

	initialRefreshToken := regRes.Tokens.RefreshToken

	// Refresh token
	newTokens, appErr := svc.RefreshToken(ctx, auth.RefreshTokenRequest{
		RefreshToken: initialRefreshToken,
	})
	if appErr != nil {
		t.Fatalf("unexpected refresh error: %v", appErr)
	}

	if newTokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if newTokens.RefreshToken == "" || newTokens.RefreshToken == initialRefreshToken {
		t.Error("expected new rotated refresh token")
	}

	// Attempting to reuse old refresh token must fail (Token rotation security)
	_, appErrReuse := svc.RefreshToken(ctx, auth.RefreshTokenRequest{
		RefreshToken: initialRefreshToken,
	})
	if appErrReuse == nil {
		t.Fatal("expected error reusing rotated refresh token, got nil")
	}
}

func TestLogout(t *testing.T) {
	svc, _ := setupService()
	ctx := context.Background()

	regRes, _ := svc.Register(ctx, auth.RegisterRequest{
		Email:     "logout@encertia.cat",
		Password:  "Password123!",
		FirstName: "Laura",
		LastName:  "Sole",
		Role:      auth.RoleStudent,
	})

	msg, appErr := svc.Logout(ctx, regRes.User.ID, regRes.Tokens.AccessToken, auth.LogoutRequest{
		RefreshToken: &regRes.Tokens.RefreshToken,
	})
	if appErr != nil {
		t.Fatalf("unexpected logout error: %v", appErr)
	}
	if msg.Message != "Sessió tancada correctament." {
		t.Errorf("expected success message, got %s", msg.Message)
	}

	// Using the revoked access token must now fail
	_, appErrAccess := svc.ValidateAccessToken(ctx, regRes.Tokens.AccessToken)
	if appErrAccess == nil {
		t.Fatal("expected error using revoked access token, got nil")
	}

	// Using the revoked refresh token must now fail
	_, appErrRefresh := svc.RefreshToken(ctx, auth.RefreshTokenRequest{
		RefreshToken: regRes.Tokens.RefreshToken,
	})
	if appErrRefresh == nil {
		t.Fatal("expected error using revoked refresh token, got nil")
	}
}

func TestGetCurrentUser(t *testing.T) {
	svc, _ := setupService()
	ctx := context.Background()

	regRes, _ := svc.Register(ctx, auth.RegisterRequest{
		Email:     "me@encertia.cat",
		Password:  "Password123!",
		FirstName: "Clara",
		LastName:  "Vila",
		Role:      auth.RoleTeacher,
	})

	userRes, appErr := svc.GetCurrentUser(ctx, regRes.User.ID)
	if appErr != nil {
		t.Fatalf("unexpected error getting current user: %v", appErr)
	}
	if userRes.User.ID != regRes.User.ID {
		t.Errorf("expected user ID %s, got %s", regRes.User.ID, userRes.User.ID)
	}

	// Non existent user
	_, notFoundErr := svc.GetCurrentUser(ctx, uuid.New())
	if notFoundErr == nil {
		t.Fatal("expected not found error, got nil")
	}
	if notFoundErr.Code != shared.ErrCodeUserNotFound {
		t.Errorf("expected USER_NOT_FOUND, got %s", notFoundErr.Code)
	}
}

func TestValidateAccessToken(t *testing.T) {
	svc, _ := setupService()
	ctx := context.Background()

	regRes, _ := svc.Register(ctx, auth.RegisterRequest{
		Email:     "jwt@encertia.cat",
		Password:  "Password123!",
		FirstName: "Marc",
		LastName:  "Serra",
		Role:      auth.RoleStudent,
	})

	claims, appErr := svc.ValidateAccessToken(ctx, regRes.Tokens.AccessToken)
	if appErr != nil {
		t.Fatalf("unexpected error validating valid token: %v", appErr)
	}
	if claims.Email != "jwt@encertia.cat" {
		t.Errorf("expected email jwt@encertia.cat, got %s", claims.Email)
	}
	if claims.Role != auth.RoleStudent {
		t.Errorf("expected role student, got %s", claims.Role)
	}

	// Invalid token string
	_, invalidErr := svc.ValidateAccessToken(ctx, "invalid.jwt.token")
	if invalidErr == nil {
		t.Fatal("expected error on invalid token string, got nil")
	}
}

func TestRegister_LanguageValidation(t *testing.T) {
	svc, _ := setupService()
	ctx := context.Background()

	// Default language 'ca'
	resDefault, appErr := svc.Register(ctx, auth.RegisterRequest{
		Email:     "lang.default@encertia.cat",
		Password:  "Password123!",
		FirstName: "Lang",
		LastName:  "Test",
		Role:      auth.RoleStudent,
	})
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if resDefault.User.Language != "ca" {
		t.Errorf("expected default language 'ca', got %s", resDefault.User.Language)
	}

	// Valid language 'es'
	resEs, appErrEs := svc.Register(ctx, auth.RegisterRequest{
		Email:     "lang.es@encertia.cat",
		Password:  "Password123!",
		FirstName: "Lang",
		LastName:  "Test",
		Role:      auth.RoleStudent,
		Language:  "es",
	})
	if appErrEs != nil {
		t.Fatalf("unexpected error: %v", appErrEs)
	}
	if resEs.User.Language != "es" {
		t.Errorf("expected language 'es', got %s", resEs.User.Language)
	}

	// Valid language upper 'EN' -> sanitized to 'en'
	resEn, appErrEn := svc.Register(ctx, auth.RegisterRequest{
		Email:     "lang.en@encertia.cat",
		Password:  "Password123!",
		FirstName: "Lang",
		LastName:  "Test",
		Role:      auth.RoleStudent,
		Language:  "EN",
	})
	if appErrEn != nil {
		t.Fatalf("unexpected error: %v", appErrEn)
	}
	if resEn.User.Language != "en" {
		t.Errorf("expected language 'en', got %s", resEn.User.Language)
	}

	// Invalid language 'fr'
	_, appErrFr := svc.Register(ctx, auth.RegisterRequest{
		Email:     "lang.fr@encertia.cat",
		Password:  "Password123!",
		FirstName: "Lang",
		LastName:  "Test",
		Role:      auth.RoleStudent,
		Language:  "fr",
	})
	if appErrFr == nil {
		t.Fatal("expected error for invalid language 'fr', got nil")
	}
	if appErrFr.StatusCode != 400 {
		t.Errorf("expected status 400 for invalid language, got %d", appErrFr.StatusCode)
	}
}

type mockDeviceValidator struct {
	validTokens map[string]uuid.UUID
}

func (m *mockDeviceValidator) ValidateAndConsumeBootstrapToken(ctx context.Context, token string, teacherID uuid.UUID) error {
	expectedTeacherID, exists := m.validTokens[token]
	if !exists || expectedTeacherID != teacherID {
		return errors.New("invalid bootstrap token")
	}
	delete(m.validTokens, token)
	return nil
}

func TestLogin_TeacherDeviceAuthorization(t *testing.T) {
	ctx := context.Background()
	repo := newMockRepository()

	teacherID := uuid.New()
	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte("TeacherPass123!"), bcrypt.DefaultCost)
	teacher := &auth.UserDB{
		ID:           teacherID,
		Email:        "teacher.auth@encertia.cat",
		PasswordHash: string(hashedPwd),
		FirstName:    "Pere",
		LastName:     "Docent",
		Role:         auth.RoleTeacher,
		Language:     "ca",
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	_ = repo.CreateUser(ctx, teacher)

	tokenVal := "dbt_validtoken123"
	devValidator := &mockDeviceValidator{
		validTokens: map[string]uuid.UUID{
			tokenVal: teacherID,
		},
	}

	// 1. Soft mode (EnforceTeacherDeviceAuth = false): teacher logs in without deviceToken -> succeeds!
	svcSoft := auth.NewService(repo, auth.Config{
		JWTSecret:                "test-secret-key-32-characters-long!",
		EnforceTeacherDeviceAuth: false,
	}, devValidator)

	res, appErr := svcSoft.Login(ctx, auth.LoginRequest{
		Email:    teacher.Email,
		Password: "TeacherPass123!",
	})
	if appErr != nil {
		t.Fatalf("expected successful login without token in soft mode, got: %v", appErr)
	}
	if res.User.Email != teacher.Email {
		t.Errorf("expected email %s, got %s", teacher.Email, res.User.Email)
	}

	// 2. Teacher logs in with valid deviceToken -> succeeds!
	resValid, appErrValid := svcSoft.Login(ctx, auth.LoginRequest{
		Email:       teacher.Email,
		Password:    "TeacherPass123!",
		DeviceToken: tokenVal,
	})
	if appErrValid != nil {
		t.Fatalf("expected successful login with valid deviceToken, got: %v", appErrValid)
	}
	if resValid.User.Email != teacher.Email {
		t.Errorf("expected email %s, got %s", teacher.Email, resValid.User.Email)
	}

	// 3. Teacher logs in with consumed/invalid deviceToken -> 403 Forbidden DEVICE_AUTHORIZATION_REQUIRED!
	_, appErrInvalid := svcSoft.Login(ctx, auth.LoginRequest{
		Email:       teacher.Email,
		Password:    "TeacherPass123!",
		DeviceToken: tokenVal, // already consumed
	})
	if appErrInvalid == nil {
		t.Fatal("expected error with consumed token, got nil")
	}
	if appErrInvalid.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", appErrInvalid.StatusCode)
	}
	if appErrInvalid.Code != shared.ErrCodeDeviceAuthRequired {
		t.Errorf("expected code DEVICE_AUTHORIZATION_REQUIRED, got %s", appErrInvalid.Code)
	}

	// 4. Strict mode (EnforceTeacherDeviceAuth = true): teacher without deviceToken -> 403 Forbidden!
	svcStrict := auth.NewService(repo, auth.Config{
		JWTSecret:                "test-secret-key-32-characters-long!",
		EnforceTeacherDeviceAuth: true,
	}, devValidator)

	_, appErrStrict := svcStrict.Login(ctx, auth.LoginRequest{
		Email:    teacher.Email,
		Password: "TeacherPass123!",
	})
	if appErrStrict == nil {
		t.Fatal("expected error without token in strict mode, got nil")
	}
	if appErrStrict.StatusCode != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", appErrStrict.StatusCode)
	}
	if appErrStrict.Code != shared.ErrCodeDeviceAuthRequired {
		t.Errorf("expected code DEVICE_AUTHORIZATION_REQUIRED, got %s", appErrStrict.Code)
	}

	// 5. Student login with or without deviceToken is unaffected
	studentHashed, _ := bcrypt.GenerateFromPassword([]byte("StudentPass123!"), bcrypt.DefaultCost)
	student := &auth.UserDB{
		ID:           uuid.New(),
		Email:        "student.auth@encertia.cat",
		PasswordHash: string(studentHashed),
		FirstName:    "Alumne",
		LastName:     "Estudiant",
		Role:         auth.RoleStudent,
		Language:     "ca",
		IsActive:     true,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	_ = repo.CreateUser(ctx, student)

	resStudent, appErrStudent := svcStrict.Login(ctx, auth.LoginRequest{
		Email:       student.Email,
		Password:    "StudentPass123!",
		DeviceToken: "random-or-empty",
	})
	if appErrStudent != nil {
		t.Fatalf("expected student login to succeed even in strict mode, got: %v", appErrStudent)
	}
	if resStudent.User.Email != student.Email {
		t.Errorf("expected student email %s, got %s", student.Email, resStudent.User.Email)
	}
}

