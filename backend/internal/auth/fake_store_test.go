package auth

import (
	"context"
	"errors"
	"sync"
	"time"
)

// fakeStore is an in-memory Store implementation for unit tests.
// It is NOT exported and lives only in _test files.
type fakeStore struct {
	mu            sync.Mutex
	users         map[string]fakeUser // keyed by email
	refreshTokens map[string]string   // userID -> tokenHash
}

type fakeUser struct {
	User
	passwordHash string
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		users:         make(map[string]fakeUser),
		refreshTokens: make(map[string]string),
	}
}

func (f *fakeStore) CreateUser(_ context.Context, email, username, passwordHash, role string) (User, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, u := range f.users {
		if u.Email == email {
			return User{}, errors.New(`ERROR: duplicate key value violates unique constraint "users_email_key"`)
		}
		if u.Username == username {
			return User{}, errors.New(`ERROR: duplicate key value violates unique constraint "users_username_key"`)
		}
	}

	u := User{
		ID:        "fake-uuid-" + email,
		Email:     email,
		Username:  username,
		Role:      role,
		IsPremium: false,
		XP:        0,
		CreatedAt: time.Now(),
	}
	f.users[email] = fakeUser{User: u, passwordHash: passwordHash}
	return u, nil
}

func (f *fakeStore) GetUserByEmail(_ context.Context, email string) (User, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	u, ok := f.users[email]
	if !ok {
		return User{}, "", errors.New("no rows in result set")
	}
	return u.User, u.passwordHash, nil
}

func (f *fakeStore) StoreRefreshToken(_ context.Context, userID, tokenHash string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.refreshTokens[userID] = tokenHash
	return nil
}

func (f *fakeStore) GetRefreshToken(_ context.Context, userID string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	h, ok := f.refreshTokens[userID]
	if !ok {
		return "", errors.New("no rows in result set")
	}
	return h, nil
}

func (f *fakeStore) DeleteRefreshToken(_ context.Context, userID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.refreshTokens, userID)
	return nil
}
