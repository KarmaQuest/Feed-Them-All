package auth

import (
	"context"
	"os"
	"testing"
)

// newTestService creates a Service backed by an in-memory fakeStore.
func newTestService(t *testing.T) (*Service, *fakeStore) {
	t.Helper()
	os.Setenv("JWT_SECRET", "test-secret-32-chars-long-padded!!")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-32-chars!!!!")
	store := newFakeStore()
	svc := NewService(store)
	return svc, store
}

// --- Register ---

func TestRegister_Success(t *testing.T) {
	svc, _ := newTestService(t)

	resp, refreshToken, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "alice@fta.dev",
		Username: "Alice",
		Password: "strongPass1",
		Roles:    []string{"feeder"},
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if refreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if resp.User.Email != "alice@fta.dev" {
		t.Errorf("expected email alice@fta.dev, got %s", resp.User.Email)
	}
	if resp.User.Role != "feeder" {
		t.Errorf("expected role feeder, got %s", resp.User.Role)
	}
}

func TestRegister_DefaultRole(t *testing.T) {
	svc, _ := newTestService(t)

	resp, _, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "bob@fta.dev",
		Username: "Bob",
		Password: "strongPass1",
		Roles:    []string{}, // empty → should default to feeder
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.User.Role != "feeder" {
		t.Errorf("expected default role feeder, got %s", resp.User.Role)
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	svc, _ := newTestService(t)

	_, _, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "weak@fta.dev",
		Username: "Weakling",
		Password: "abc", // too short
	})

	if err != ErrWeakPassword {
		t.Errorf("expected ErrWeakPassword, got %v", err)
	}
}

func TestRegister_InvalidRole(t *testing.T) {
	svc, _ := newTestService(t)

	_, _, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "bad@fta.dev",
		Username: "BadRole",
		Password: "strongPass1",
		Roles:    []string{"admin"}, // invalid
	})

	if err != ErrInvalidRole {
		t.Errorf("expected ErrInvalidRole, got %v", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc, _ := newTestService(t)

	req := RegisterRequest{
		Email:    "dup@fta.dev",
		Username: "First",
		Password: "strongPass1",
		Roles:    []string{"feeder"},
	}
	if _, _, err := svc.Register(context.Background(), req); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	req.Username = "Second"
	_, _, err := svc.Register(context.Background(), req)
	if err != ErrEmailTaken {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	svc, _ := newTestService(t)

	req := RegisterRequest{
		Email:    "first@fta.dev",
		Username: "SharedName",
		Password: "strongPass1",
		Roles:    []string{"feeder"},
	}
	if _, _, err := svc.Register(context.Background(), req); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	req.Email = "second@fta.dev"
	_, _, err := svc.Register(context.Background(), req)
	if err != ErrUsernameTaken {
		t.Errorf("expected ErrUsernameTaken, got %v", err)
	}
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	svc, _ := newTestService(t)

	// Register first
	if _, _, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "login@fta.dev",
		Username: "LoginUser",
		Password: "correctPass1",
		Roles:    []string{"giver"},
	}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	resp, refreshToken, err := svc.Login(context.Background(), LoginRequest{
		Email:    "login@fta.dev",
		Password: "correctPass1",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected access token")
	}
	if refreshToken == "" {
		t.Error("expected refresh token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, _ := newTestService(t)

	if _, _, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "pw@fta.dev",
		Username: "PwUser",
		Password: "correctPass1",
		Roles:    []string{"feeder"},
	}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, _, err := svc.Login(context.Background(), LoginRequest{
		Email:    "pw@fta.dev",
		Password: "wrongPass99",
	})
	if err != ErrInvalidCreds {
		t.Errorf("expected ErrInvalidCreds, got %v", err)
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	svc, _ := newTestService(t)

	_, _, err := svc.Login(context.Background(), LoginRequest{
		Email:    "nobody@fta.dev",
		Password: "anything",
	})
	if err != ErrInvalidCreds {
		t.Errorf("expected ErrInvalidCreds, got %v", err)
	}
}

// --- Refresh ---

func TestRefresh_Success(t *testing.T) {
	svc, _ := newTestService(t)

	if _, _, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "refresh@fta.dev",
		Username: "RefreshUser",
		Password: "refreshPass1",
		Roles:    []string{"feeder"},
	}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	_, refreshToken, err := svc.Login(context.Background(), LoginRequest{
		Email:    "refresh@fta.dev",
		Password: "refreshPass1",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	resp, newRefreshToken, err := svc.Refresh(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected new access token")
	}
	if newRefreshToken == refreshToken {
		t.Error("expected a rotated refresh token, got the same one")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	svc, _ := newTestService(t)

	_, _, err := svc.Refresh(context.Background(), "not.a.valid.token")
	if err == nil {
		t.Error("expected error for invalid refresh token")
	}
}

// --- Logout ---

func TestLogout_ClearsRefreshToken(t *testing.T) {
	svc, store := newTestService(t)

	if _, _, err := svc.Register(context.Background(), RegisterRequest{
		Email:    "logout@fta.dev",
		Username: "LogoutUser",
		Password: "logoutPass1",
		Roles:    []string{"feeder"},
	}); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	resp, _, err := svc.Login(context.Background(), LoginRequest{
		Email:    "logout@fta.dev",
		Password: "logoutPass1",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if err := svc.Logout(context.Background(), resp.User.ID); err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	// Token should be gone from store
	_, err = store.GetRefreshToken(context.Background(), resp.User.ID)
	if err == nil {
		t.Error("expected error fetching deleted refresh token")
	}
}
