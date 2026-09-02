package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/encertia/backend/internal/auth"
	"github.com/encertia/backend/internal/shared"
	"github.com/google/uuid"
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
