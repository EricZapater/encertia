package user_test

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/encertia/backend/internal/user"
	"github.com/google/uuid"
)

type mockRepository struct {
	mu           sync.RWMutex
	users        map[string]*user.UserDB    // email -> user
	usersByID    map[uuid.UUID]*user.UserDB // id -> user
	revokedUsers map[uuid.UUID]time.Time
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		users:        make(map[string]*user.UserDB),
		usersByID:    make(map[uuid.UUID]*user.UserDB),
		revokedUsers: make(map[uuid.UUID]time.Time),
	}
}

func (m *mockRepository) ListUsers(ctx context.Context, filter user.ListUsersFilter) ([]user.UserDB, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filtered []user.UserDB
	searchLower := strings.ToLower(strings.TrimSpace(filter.Search))

	for _, u := range m.usersByID {
		if u.DeletedAt != nil {
			continue
		}

		// Role scoping
		if filter.ActorRole == string(user.RoleTeacher) {
			if u.Role != user.RoleStudent {
				continue
			}
		} else if filter.Role != "" {
			if string(u.Role) != filter.Role {
				continue
			}
		}

		// Status filtering
		switch filter.Status {
		case "inactive":
			if u.IsActive {
				continue
			}
		case "all":
			// include all
		default: // "active"
			if !u.IsActive {
				continue
			}
		}

		// Search filtering
		if searchLower != "" {
			fn := strings.ToLower(u.FirstName)
			ln := strings.ToLower(u.LastName)
			em := strings.ToLower(u.Email)
			if !strings.Contains(fn, searchLower) &&
				!strings.Contains(ln, searchLower) &&
				!strings.Contains(em, searchLower) {
				continue
			}
		}

		filtered = append(filtered, *u)
	}

	totalCount := len(filtered)
	limit := filter.PageSize
	if limit <= 0 {
		limit = 20
	}
	page := filter.Page
	if page <= 0 {
		page = 1
	}

	start := (page - 1) * limit
	if start >= totalCount {
		return []user.UserDB{}, totalCount, nil
	}

	end := start + limit
	if end > totalCount {
		end = totalCount
	}

	return filtered[start:end], totalCount, nil
}

func (m *mockRepository) CreateUser(ctx context.Context, u *user.UserDB) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if existing, exists := m.users[strings.ToLower(u.Email)]; exists && existing.DeletedAt == nil {
		return user.ErrEmailAlreadyInUse
	}

	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	now := time.Now().UTC()
	u.CreatedAt = now
	u.UpdatedAt = now

	m.users[strings.ToLower(u.Email)] = u
	m.usersByID[u.ID] = u
	return nil
}

func (m *mockRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*user.UserDB, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	u, exists := m.usersByID[id]
	if !exists || u.DeletedAt != nil {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (m *mockRepository) GetUserByEmail(ctx context.Context, email string) (*user.UserDB, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	u, exists := m.users[strings.ToLower(email)]
	if !exists || u.DeletedAt != nil {
		return nil, user.ErrUserNotFound
	}
	return u, nil
}

func (m *mockRepository) UpdateUser(ctx context.Context, u *user.UserDB) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, exists := m.usersByID[u.ID]
	if !exists || existing.DeletedAt != nil {
		return user.ErrUserNotFound
	}

	// Update email key if changed
	if strings.ToLower(existing.Email) != strings.ToLower(u.Email) {
		delete(m.users, strings.ToLower(existing.Email))
		m.users[strings.ToLower(u.Email)] = u
	}

	u.UpdatedAt = time.Now().UTC()
	m.usersByID[u.ID] = u
	return nil
}

func (m *mockRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	u, exists := m.usersByID[id]
	if !exists || u.DeletedAt != nil {
		return user.ErrUserNotFound
	}

	u.PasswordHash = passwordHash
	u.UpdatedAt = time.Now().UTC()
	return nil
}

func (m *mockRepository) SoftDeleteUser(ctx context.Context, id uuid.UUID, deletedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	u, exists := m.usersByID[id]
	if !exists || u.DeletedAt != nil {
		return user.ErrUserNotFound
	}

	u.DeletedAt = &deletedAt
	u.UpdatedAt = deletedAt
	return nil
}

func (m *mockRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID, revokedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.revokedUsers[userID] = revokedAt
	return nil
}
