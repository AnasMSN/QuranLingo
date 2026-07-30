package handler

import (
	"errors"
	"net/http"

	"quranlingo/backend/internal/httpx"
	"quranlingo/backend/internal/middleware"
	"quranlingo/backend/internal/models"
	"quranlingo/backend/internal/repository"
	"quranlingo/backend/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
	db   repository.Querier
}

func NewAuthHandler(auth *service.AuthService, db repository.Querier) *AuthHandler {
	return &AuthHandler{auth: auth, db: db}
}

type userResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
	TotalXP       int    `json:"total_xp"`
	Hearts        int    `json:"hearts"`
	CurrentStreak int    `json:"current_streak"`
	LongestStreak int    `json:"longest_streak"`
}

func toUserResponse(u *models.User) userResponse {
	return userResponse{
		ID:            u.ID,
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		TotalXP:       u.TotalXP,
		Hearts:        u.Hearts,
		CurrentStreak: u.CurrentStreak,
		LongestStreak: u.LongestStreak,
	}
}

type authResponse struct {
	User   userResponse       `json:"user"`
	Tokens *service.TokenPair `json:"tokens"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Email == "" || body.Password == "" || body.DisplayName == "" {
		httpx.Error(w, http.StatusBadRequest, "email, password, and display_name are required")
		return
	}

	user, tokens, err := h.auth.Register(r.Context(), body.Email, body.Password, body.DisplayName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailTaken):
			httpx.Error(w, http.StatusConflict, err.Error())
		default:
			httpx.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	httpx.JSON(w, http.StatusCreated, authResponse{User: toUserResponse(user), Tokens: tokens})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, tokens, err := h.auth.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	httpx.JSON(w, http.StatusOK, authResponse{User: toUserResponse(user), Tokens: tokens})
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil || body.RefreshToken == "" {
		httpx.Error(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	tokens, err := h.auth.Refresh(r.Context(), body.RefreshToken)
	if err != nil {
		httpx.Error(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return
	}

	httpx.JSON(w, http.StatusOK, map[string]*service.TokenPair{"tokens": tokens})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserID(r.Context())

	user, err := repository.GetUserByID(r.Context(), h.db, userID)
	if err != nil {
		httpx.Error(w, http.StatusNotFound, "user not found")
		return
	}

	httpx.JSON(w, http.StatusOK, toUserResponse(user))
}
