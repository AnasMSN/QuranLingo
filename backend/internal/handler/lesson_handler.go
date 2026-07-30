package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"quranlingo/backend/internal/httpx"
	"quranlingo/backend/internal/middleware"
	"quranlingo/backend/internal/repository"
	"quranlingo/backend/internal/service"
)

type LessonHandler struct {
	lessons *service.LessonService
}

func NewLessonHandler(lessons *service.LessonService) *LessonHandler {
	return &LessonHandler{lessons: lessons}
}

func (h *LessonHandler) Submit(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserID(r.Context())
	lessonID := chi.URLParam(r, "id")

	var body struct {
		IdempotencyKey string                  `json:"idempotency_key"`
		Answers        []service.AnswerInput   `json:"answers"`
	}
	if err := httpx.DecodeJSON(r, &body); err != nil {
		httpx.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.lessons.Submit(r.Context(), userID, lessonID, body.IdempotencyKey, body.Answers)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			httpx.Error(w, http.StatusNotFound, "lesson not found")
		case errors.Is(err, service.ErrNoHeartsRemaining):
			httpx.Error(w, http.StatusForbidden, "no hearts remaining, wait for refill")
		case errors.Is(err, service.ErrIdempotencyKeyReq):
			httpx.Error(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrLessonHasNoContent):
			httpx.Error(w, http.StatusUnprocessableEntity, err.Error())
		default:
			httpx.Error(w, http.StatusInternalServerError, "failed to submit lesson")
		}
		return
	}

	httpx.JSON(w, http.StatusOK, result)
}
