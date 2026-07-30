package handler

import (
	"net/http"

	"quranlingo/backend/internal/httpx"
	"quranlingo/backend/internal/service"
)

type LeaderboardHandler struct {
	leaderboard *service.LeaderboardService
}

func NewLeaderboardHandler(leaderboard *service.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{leaderboard: leaderboard}
}

func (h *LeaderboardHandler) Weekly(w http.ResponseWriter, r *http.Request) {
	entries, err := h.leaderboard.Weekly(r.Context(), 50)
	if err != nil {
		httpx.Error(w, http.StatusInternalServerError, "failed to load leaderboard")
		return
	}
	httpx.JSON(w, http.StatusOK, map[string]any{"entries": entries})
}
