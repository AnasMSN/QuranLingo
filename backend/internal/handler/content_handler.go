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

type ContentHandler struct {
	content *service.ContentService
}

func NewContentHandler(content *service.ContentService) *ContentHandler {
	return &ContentHandler{content: content}
}

func (h *ContentHandler) GetCourse(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserID(r.Context())
	code := chi.URLParam(r, "code")

	tree, err := h.content.GetCourseTree(r.Context(), userID, code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			httpx.Error(w, http.StatusNotFound, "course not found")
			return
		}
		httpx.Error(w, http.StatusInternalServerError, "failed to load course")
		return
	}

	httpx.JSON(w, http.StatusOK, tree)
}

func (h *ContentHandler) GetLesson(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.UserID(r.Context())
	lessonID := chi.URLParam(r, "id")

	detail, err := h.content.GetLessonDetail(r.Context(), userID, lessonID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			httpx.Error(w, http.StatusNotFound, "lesson not found")
		case errors.Is(err, service.ErrLessonLocked):
			httpx.Error(w, http.StatusForbidden, "lesson is locked")
		default:
			httpx.Error(w, http.StatusInternalServerError, "failed to load lesson")
		}
		return
	}

	httpx.JSON(w, http.StatusOK, detail)
}
