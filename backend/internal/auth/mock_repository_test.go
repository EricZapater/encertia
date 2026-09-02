package auth_test

import (
	"context"
	"sync"
	"time"

	"github.com/encertia/backend/internal/auth"
	"github.com/google/uuid"
)

type mockRepository struct {
	mu                  sync.RWMutex
	users               map[string]*auth.UserDB         // email -> user
	usersByID           map[uuid.UUID]*auth.UserDB      // id -> user
	refreshTokens       map[string]*auth.RefreshTokenDB // tokenHash -> rt
	revokedAccessTokens map[string]bool                 // tokenHash -> isRevoked
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		users:               make(map[string]*auth.UserDB),
		usersByID:           make(map[uuid.UUID]*auth.UserDB),
		refreshTokens:       make(map[string]*auth.RefreshTokenDB),
		revokedAccessTokens: make(map[string]bool),
	}
}

func (m *mockRepository) CreateUser(ctx context.Context, user *auth.UserDB) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[user.Email]; exists {
		return auth.ErrEmailAlreadyInUse
	}

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	m.users[user.Email] = user
	m.usersByID[user.ID] = user
	return nil
}

func (m *mockRepository) GetUserByEmail(ctx context.Context, email string) (*auth.UserDB, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[email]
	if !exists || user.DeletedAt != nil {
		return nil, auth.ErrUserNotFound
	}
	return user, nil
}

func (m *mockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*auth.UserDB, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.usersByID[id]
	if !exists || user.DeletedAt != nil {
		return nil, auth.ErrUserNotFound
	}
	return user, nil
}

func (m *mockRepository) SaveRefreshToken(ctx context.Context, rt *auth.RefreshTokenDB) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if rt.ID == uuid.Nil {
		rt.ID = uuid.New()
	}
	now := time.Now().UTC()
	rt.CreatedAt = now
	rt.UpdatedAt = now

	m.refreshTokens[rt.TokenHash] = rt
	return nil
}

func (m *mockRepository) GetRefreshToken(ctx context.Context, tokenHash string) (*auth.RefreshTokenDB, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rt, exists := m.refreshTokens[tokenHash]
	if !exists || rt.DeletedAt != nil {
		return nil, auth.ErrTokenNotFound
	}
	return rt, nil
}

func (m *mockRepository) RevokeRefreshToken(ctx context.Context, tokenHash string, revokedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rt, exists := m.refreshTokens[tokenHash]
	if !exists || rt.DeletedAt != nil {
		return auth.ErrTokenNotFound
	}
	rt.RevokedAt = &revokedAt
	rt.UpdatedAt = revokedAt
	return nil
}

func (m *mockRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, rt := range m.refreshTokens {
		if rt.UserID == userID && rt.RevokedAt == nil && rt.DeletedAt == nil {
			rt.RevokedAt = &revokedAt
			rt.UpdatedAt = revokedAt
		}
	}
	return nil
}

func (m *mockRepository) RevokeAccessToken(ctx context.Context, tokenHash string, userID uuid.UUID, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.revokedAccessTokens[tokenHash] = true
	return nil
}

func (m *mockRepository) IsAccessTokenRevoked(ctx context.Context, tokenHash string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.revokedAccessTokens[tokenHash], nil
}
