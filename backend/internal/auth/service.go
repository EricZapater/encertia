package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/encertia/backend/internal/shared"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	DefaultAccessTokenDuration  = 15 * time.Minute
	DefaultRefreshTokenDuration = 7 * 24 * time.Hour
	AccessTokenTokenType        = "Bearer"
)

// Service defines the business logic operations for the Auth domain.
type Service interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, *shared.AppError)
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, *shared.AppError)
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (*TokenPair, *shared.AppError)
	Logout(ctx context.Context, userID uuid.UUID, req LogoutRequest) (*MessageResponse, *shared.AppError)
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*UserResponse, *shared.AppError)
	ValidateAccessToken(tokenString string) (*JWTClaims, *shared.AppError)
}

type Config struct {
	JWTSecret            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

type authService struct {
	repo Repository
	cfg  Config
}

// NewService creates a new instance of AuthService.
func NewService(repo Repository, cfg Config) Service {
	if cfg.AccessTokenDuration == 0 {
		cfg.AccessTokenDuration = DefaultAccessTokenDuration
	}
	if cfg.RefreshTokenDuration == 0 {
		cfg.RefreshTokenDuration = DefaultRefreshTokenDuration
	}
	if cfg.Issuer == "" {
		cfg.Issuer = "encertia-auth"
	}
	return &authService{
		repo: repo,
		cfg:  cfg,
	}
}

func (s *authService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, *shared.AppError) {
	// 1. Validation
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)

	if req.Email == "" {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El correu electrònic és obligatori.", map[string]interface{}{"field": "email"})
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El format del correu electrònic no és vàlid.", map[string]interface{}{"field": "email"})
	}
	if len(req.Password) < 8 {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "La contrasenya ha de tenir com a mínim 8 caràcters.", map[string]interface{}{"field": "password"})
	}
	if req.FirstName == "" {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El nom és obligatori.", map[string]interface{}{"field": "firstName"})
	}
	if req.LastName == "" {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "Els cognoms són obligatoris.", map[string]interface{}{"field": "lastName"})
	}

	// 2. Check if user already exists
	existingUser, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, shared.ErrInternal(err)
	}
	if existingUser != nil {
		return nil, shared.ErrConflict(shared.ErrCodeEmailAlreadyExists, "El correu electrònic ja està registrat.")
	}

	// 3. Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	// 4. Save user (public registration always creates students with active status)
	userDB := &UserDB{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         RoleStudent,
		IsActive:     true,
	}

	if err := s.repo.CreateUser(ctx, userDB); err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return nil, shared.ErrConflict(shared.ErrCodeEmailAlreadyExists, "El correu electrònic ja està registrat.")
		}
		return nil, shared.ErrInternal(err)
	}

	// 5. Generate tokens
	tokenPair, appErr := s.generateAndStoreTokens(ctx, userDB)
	if appErr != nil {
		return nil, appErr
	}

	return &AuthResponse{
		User:   userDB.ToUser(),
		Tokens: *tokenPair,
	}, nil
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, *shared.AppError) {
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "Correu electrònic i contrasenya són obligatoris.", nil)
	}

	userDB, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, shared.ErrUnauthorized(shared.ErrCodeInvalidCredentials, "El correu o la contrasenya són incorrectes.")
		}
		return nil, shared.ErrInternal(err)
	}

	if !userDB.IsActive {
		return nil, shared.ErrUnauthorized(shared.ErrCodeUnauthorized, "El compte d'usuari està desactivat.")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userDB.PasswordHash), []byte(req.Password)); err != nil {
		return nil, shared.ErrUnauthorized(shared.ErrCodeInvalidCredentials, "El correu o la contrasenya són incorrectes.")
	}

	tokenPair, appErr := s.generateAndStoreTokens(ctx, userDB)
	if appErr != nil {
		return nil, appErr
	}

	return &AuthResponse{
		User:   userDB.ToUser(),
		Tokens: *tokenPair,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*TokenPair, *shared.AppError) {
	req.RefreshToken = strings.TrimSpace(req.RefreshToken)
	if req.RefreshToken == "" {
		return nil, shared.ErrBadRequest(shared.ErrCodeValidation, "El refresh token és obligatori.", map[string]interface{}{"field": "refreshToken"})
	}

	tokenHash := hashToken(req.RefreshToken)
	storedRT, err := s.repo.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrTokenNotFound) {
			return nil, shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "Refresh token no vàlid o inexistent.")
		}
		return nil, shared.ErrInternal(err)
	}

	if storedRT.RevokedAt != nil {
		return nil, shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "El refresh token ha estat revocat.")
	}

	if time.Now().UTC().After(storedRT.ExpiresAt) {
		return nil, shared.ErrUnauthorized(shared.ErrCodeTokenExpired, "El refresh token ha expirat.")
	}

	userDB, err := s.repo.GetUserByID(ctx, storedRT.UserID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "Usuari associat no trobat.")
		}
		return nil, shared.ErrInternal(err)
	}

	// Token rotation: Revoke the old refresh token
	now := time.Now().UTC()
	_ = s.repo.RevokeRefreshToken(ctx, tokenHash, now)

	// Issue new token pair
	tokenPair, appErr := s.generateAndStoreTokens(ctx, userDB)
	if appErr != nil {
		return nil, appErr
	}

	return tokenPair, nil
}

func (s *authService) Logout(ctx context.Context, userID uuid.UUID, req LogoutRequest) (*MessageResponse, *shared.AppError) {
	now := time.Now().UTC()

	if req.RefreshToken != nil && strings.TrimSpace(*req.RefreshToken) != "" {
		tokenHash := hashToken(strings.TrimSpace(*req.RefreshToken))
		if err := s.repo.RevokeRefreshToken(ctx, tokenHash, now); err != nil {
			return nil, shared.ErrInternal(err)
		}
	} else {
		if err := s.repo.RevokeAllUserRefreshTokens(ctx, userID, now); err != nil {
			return nil, shared.ErrInternal(err)
		}
	}

	return &MessageResponse{
		Message: "Sessió tancada correctament.",
	}, nil
}

func (s *authService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*UserResponse, *shared.AppError) {
	userDB, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, shared.ErrNotFound(shared.ErrCodeUserNotFound, "Usuari no trobat.")
		}
		return nil, shared.ErrInternal(err)
	}

	return &UserResponse{
		User: userDB.ToUser(),
	}, nil
}

func (s *authService) ValidateAccessToken(tokenString string) (*JWTClaims, *shared.AppError) {
	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("mètode de signatura inesperat: %v", t.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, shared.ErrUnauthorized(shared.ErrCodeTokenExpired, "El token d'accés ha expirat.")
		}
		return nil, shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "El token d'accés és invàlid.")
	}

	if !token.Valid {
		return nil, shared.ErrUnauthorized(shared.ErrCodeInvalidToken, "El token d'accés és invàlid.")
	}

	return claims, nil
}

func (s *authService) generateAndStoreTokens(ctx context.Context, userDB *UserDB) (*TokenPair, *shared.AppError) {
	now := time.Now().UTC()
	accessExpiry := now.Add(s.cfg.AccessTokenDuration)
	refreshExpiry := now.Add(s.cfg.RefreshTokenDuration)

	// Access Token
	claims := JWTClaims{
		UserID: userDB.ID.String(),
		Email:  userDB.Email,
		Role:   userDB.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userDB.ID.String(),
			Issuer:    s.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedAccessToken, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	// Refresh Token (cryptographically random string)
	rawRefreshToken, err := generateRandomToken(32)
	if err != nil {
		return nil, shared.ErrInternal(err)
	}

	rtDB := &RefreshTokenDB{
		ID:        uuid.New(),
		UserID:    userDB.ID,
		TokenHash: hashToken(rawRefreshToken),
		ExpiresAt: refreshExpiry,
	}

	if err := s.repo.SaveRefreshToken(ctx, rtDB); err != nil {
		return nil, shared.ErrInternal(err)
	}

	return &TokenPair{
		AccessToken:  signedAccessToken,
		RefreshToken: rawRefreshToken,
		TokenType:    AccessTokenTokenType,
		ExpiresIn:    int(s.cfg.AccessTokenDuration.Seconds()),
	}, nil
}

func generateRandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
