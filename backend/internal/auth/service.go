package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Sentinel errors
var (
	ErrEmailTaken    = errors.New("email already in use")
	ErrUsernameTaken = errors.New("username already in use")
	ErrInvalidCreds  = errors.New("invalid email or password")
	ErrInvalidRole   = errors.New("invalid role: must be feeder, giver or association")
	ErrWeakPassword  = errors.New("password must be at least 8 characters")
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

// Service holds business logic for authentication.
type Service struct {
	repo           *Repository
	jwtSecret      []byte
	jwtRefreshSecret []byte
}

func NewService(repo *Repository) *Service {
	secret := os.Getenv("JWT_SECRET")
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if secret == "" || refreshSecret == "" {
		panic("JWT_SECRET and JWT_REFRESH_SECRET must be set")
	}
	return &Service{
		repo:             repo,
		jwtSecret:        []byte(secret),
		jwtRefreshSecret: []byte(refreshSecret),
	}
}

// Register creates a new user and returns tokens.
func (s *Service) Register(ctx context.Context, req RegisterRequest) (TokenResponse, string, error) {
	// Validate role
	role := strings.ToLower(req.Role)
	if role == "" {
		role = "feeder"
	}
	if role != "feeder" && role != "giver" && role != "association" {
		return TokenResponse{}, "", ErrInvalidRole
	}

	// Validate password length
	if len(req.Password) < 8 {
		return TokenResponse{}, "", ErrWeakPassword
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return TokenResponse{}, "", fmt.Errorf("auth.Register hash: %w", err)
	}

	// Insert user
	user, err := s.repo.CreateUser(ctx, req.Email, req.Username, string(hash), role)
	if err != nil {
		// Map DB constraint errors to friendly sentinels
		msg := err.Error()
		if strings.Contains(msg, "users_email_key") {
			return TokenResponse{}, "", ErrEmailTaken
		}
		if strings.Contains(msg, "users_username_key") {
			return TokenResponse{}, "", ErrUsernameTaken
		}
		return TokenResponse{}, "", fmt.Errorf("auth.Register insert: %w", err)
	}

	return s.issueTokens(ctx, user)
}

// Login verifies credentials and returns tokens.
func (s *Service) Login(ctx context.Context, req LoginRequest) (TokenResponse, string, error) {
	user, hash, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal whether the email exists
		return TokenResponse{}, "", ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		return TokenResponse{}, "", ErrInvalidCreds
	}

	return s.issueTokens(ctx, user)
}

// Refresh validates a refresh token and issues new tokens.
func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (TokenResponse, string, error) {
	claims, err := s.parseToken(rawRefreshToken, s.jwtRefreshSecret)
	if err != nil {
		return TokenResponse{}, "", fmt.Errorf("auth.Refresh: %w", err)
	}

	userID, _ := claims.GetSubject()

	// Verify stored hash matches
	storedHash, err := s.repo.GetRefreshToken(ctx, userID)
	if err != nil {
		return TokenResponse{}, "", ErrInvalidCreds
	}
	incoming := hashToken(rawRefreshToken)
	if incoming != storedHash {
		return TokenResponse{}, "", ErrInvalidCreds
	}

	// Fetch fresh user data
	user, _, err := s.repo.GetUserByEmail(ctx, claims["email"].(string))
	if err != nil {
		return TokenResponse{}, "", fmt.Errorf("auth.Refresh fetch user: %w", err)
	}

	return s.issueTokens(ctx, user)
}

// Logout removes the stored refresh token.
func (s *Service) Logout(ctx context.Context, userID string) error {
	return s.repo.DeleteRefreshToken(ctx, userID)
}

// ValidateAccessToken parses and validates an access token, returning the user ID.
func (s *Service) ValidateAccessToken(tokenStr string) (string, error) {
	claims, err := s.parseToken(tokenStr, s.jwtSecret)
	if err != nil {
		return "", err
	}
	userID, err := claims.GetSubject()
	if err != nil {
		return "", fmt.Errorf("auth.ValidateAccessToken: missing sub: %w", err)
	}
	return userID, nil
}

// --- internal helpers ---

func (s *Service) issueTokens(ctx context.Context, user User) (TokenResponse, string, error) {
	accessToken, err := s.signToken(user, accessTokenTTL, s.jwtSecret)
	if err != nil {
		return TokenResponse{}, "", fmt.Errorf("auth.issueTokens access: %w", err)
	}

	refreshToken, err := s.signRefreshToken(user, refreshTokenTTL, s.jwtRefreshSecret)
	if err != nil {
		return TokenResponse{}, "", fmt.Errorf("auth.issueTokens refresh: %w", err)
	}

	// Store hashed refresh token
	if err := s.repo.StoreRefreshToken(ctx, user.ID, hashToken(refreshToken)); err != nil {
		return TokenResponse{}, "", err
	}

	return TokenResponse{AccessToken: accessToken, User: user}, refreshToken, nil
}

func (s *Service) signToken(user User, ttl time.Duration, secret []byte) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"role":  user.Role,
		"iat":   now.Unix(),
		"exp":   now.Add(ttl).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

func (s *Service) signRefreshToken(user User, ttl time.Duration, secret []byte) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"iat":   now.Unix(),
		"exp":   now.Add(ttl).Unix(),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

func (s *Service) parseToken(tokenStr string, secret []byte) (jwt.MapClaims, error) {
	t, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil || !t.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// hashToken returns a SHA-256 hex string of the raw token for safe storage.
func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h)
}
