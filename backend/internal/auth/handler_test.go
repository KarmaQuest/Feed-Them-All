package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func newTestHandler(t *testing.T) (*Handler, *fakeStore) {
	t.Helper()
	os.Setenv("JWT_SECRET", "test-secret-32-chars-long-padded!!")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-32-chars!!!!")
	store := newFakeStore()
	svc := NewService(store)
	return NewHandler(svc), store
}

func postJSON(t *testing.T, handler http.HandlerFunc, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler(rr, req)
	return rr
}

// --- POST /auth/register ---

func TestHandler_Register_Created(t *testing.T) {
	h, _ := newTestHandler(t)

	rr := postJSON(t, h.Register, "/auth/register", map[string]string{
		"email":    "handler@fta.dev",
		"username": "HandlerUser",
		"password": "handlerPass1",
		"role":     "feeder",
	})

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", rr.Code, rr.Body.String())
	}

	var resp TokenResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected access_token in response")
	}
	if resp.User.Email != "handler@fta.dev" {
		t.Errorf("expected email handler@fta.dev, got %s", resp.User.Email)
	}
}

func TestHandler_Register_MissingFields(t *testing.T) {
	h, _ := newTestHandler(t)

	rr := postJSON(t, h.Register, "/auth/register", map[string]string{
		"email": "partial@fta.dev",
		// missing username and password
	})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandler_Register_InvalidBody(t *testing.T) {
	h, _ := newTestHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", rr.Code)
	}
}

func TestHandler_Register_DuplicateEmail(t *testing.T) {
	h, _ := newTestHandler(t)

	body := map[string]string{
		"email":    "dup@fta.dev",
		"username": "DupUser",
		"password": "strongPass1",
		"role":     "feeder",
	}
	postJSON(t, h.Register, "/auth/register", body) // first — should succeed

	body["username"] = "DupUser2"
	rr := postJSON(t, h.Register, "/auth/register", body) // same email
	if rr.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict, got %d", rr.Code)
	}
}

func TestHandler_Register_WeakPassword(t *testing.T) {
	h, _ := newTestHandler(t)

	rr := postJSON(t, h.Register, "/auth/register", map[string]string{
		"email":    "weak@fta.dev",
		"username": "Weakling",
		"password": "abc",
	})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for weak password, got %d", rr.Code)
	}
}

// --- POST /auth/login ---

func TestHandler_Login_Success(t *testing.T) {
	h, _ := newTestHandler(t)

	// Pre-register
	postJSON(t, h.Register, "/auth/register", map[string]string{
		"email": "login@fta.dev", "username": "LoginUser",
		"password": "loginPass1", "role": "feeder",
	})

	rr := postJSON(t, h.Login, "/auth/login", map[string]string{
		"email":    "login@fta.dev",
		"password": "loginPass1",
	})

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", rr.Code, rr.Body.String())
	}

	// Verify refresh_token cookie is set
	cookies := rr.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "refresh_token" && c.HttpOnly {
			found = true
		}
	}
	if !found {
		t.Error("expected HttpOnly refresh_token cookie in response")
	}
}

func TestHandler_Login_WrongPassword(t *testing.T) {
	h, _ := newTestHandler(t)

	postJSON(t, h.Register, "/auth/register", map[string]string{
		"email": "pw@fta.dev", "username": "PwUser",
		"password": "correctPass1", "role": "feeder",
	})

	rr := postJSON(t, h.Login, "/auth/login", map[string]string{
		"email":    "pw@fta.dev",
		"password": "wrongPass99",
	})

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestHandler_Login_UnknownEmail(t *testing.T) {
	h, _ := newTestHandler(t)

	rr := postJSON(t, h.Login, "/auth/login", map[string]string{
		"email":    "nobody@fta.dev",
		"password": "anything",
	})

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}
