package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"quranlingo/backend/internal/config"
	"quranlingo/backend/internal/models"
	"quranlingo/backend/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AuthService struct {
	db            *pgxpool.Pool
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

func NewAuthService(db *pgxpool.Pool, cfg *config.Config) *AuthService {
	return &AuthService{
		db:            db,
		accessSecret:  []byte(cfg.JWTAccessSecret),
		refreshSecret: []byte(cfg.JWTRefreshSecret),
		accessTTL:     cfg.AccessTTL,
		refreshTTL:    cfg.RefreshTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, displayName string) (*models.User, *TokenPair, error) {
	email = normalizeEmail(email)
	if len(password) < 8 {
		return nil, nil, errors.New("password must be at least 8 characters")
	}

	if _, err := repository.GetUserByEmail(ctx, s.db, email); err == nil {
		return nil, nil, ErrEmailTaken
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}

	user, err := repository.CreateUser(ctx, s.db, email, string(hash), displayName)
	if err != nil {
		return nil, nil, err
	}

	tokens, err := s.issueTokenPair(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}
	return user, tokens, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*models.User, *TokenPair, error) {
	email = normalizeEmail(email)
	user, err := repository.GetUserByEmail(ctx, s.db, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	tokens, err := s.issueTokenPair(ctx, user.ID)
	if err != nil {
		return nil, nil, err
	}
	return user, tokens, nil
}

// Refresh rotates a refresh token: the presented token is revoked and a new pair is issued.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	hash := hashToken(refreshToken)

	rt, err := repository.GetRefreshTokenByHash(ctx, s.db, hash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, err
	}
	if rt.RevokedAt != nil || time.Now().After(rt.ExpiresAt) {
		return nil, ErrInvalidToken
	}

	if err := repository.RevokeRefreshToken(ctx, s.db, hash); err != nil {
		return nil, err
	}

	return s.issueTokenPair(ctx, rt.UserID)
}

func (s *AuthService) issueTokenPair(ctx context.Context, userID string) (*TokenPair, error) {
	access, err := s.signAccessToken(userID)
	if err != nil {
		return nil, err
	}

	refresh, err := generateOpaqueToken()
	if err != nil {
		return nil, err
	}
	if err := repository.CreateRefreshToken(ctx, s.db, userID, hashToken(refresh), time.Now().Add(s.refreshTTL)); err != nil {
		return nil, err
	}

	return &TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}

func (s *AuthService) signAccessToken(userID string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTTL)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.accessSecret)
}

// ParseAccessToken validates a JWT and returns the subject (user ID).
func (s *AuthService) ParseAccessToken(tokenString string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.accessSecret, nil
	})
	if err != nil || !token.Valid {
		return "", ErrInvalidToken
	}
	return claims.Subject, nil
}

func generateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
