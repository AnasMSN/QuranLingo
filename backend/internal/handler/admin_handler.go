package handler

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"quranlingo/backend/internal/models"
	"quranlingo/backend/internal/service"
)

//go:embed templates/admin.html
var adminTemplateFS embed.FS

var adminTemplate = template.Must(
	template.New("admin.html").Funcs(template.FuncMap{
		"formatTime": func(t *time.Time) string {
			if t == nil {
				return "—"
			}
			return t.Format("2006-01-02 15:04 MST")
		},
	}).ParseFS(adminTemplateFS, "templates/admin.html"),
)

// AdminHandler serves the operator-only, server-rendered HTML admin panel.
// It deliberately does not use the httpx JSON helpers — this surface speaks
// HTML, not the mobile app's JSON API.
type AdminHandler struct {
	admin *service.AdminService
}

func NewAdminHandler(admin *service.AdminService) *AdminHandler {
	return &AdminHandler{admin: admin}
}

type adminPageData struct {
	Users     []models.User
	MaxHearts int
	Message   string
}

func (h *AdminHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	users, err := h.admin.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "failed to load users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = adminTemplate.Execute(w, adminPageData{
		Users:     users,
		MaxHearts: service.MaxHearts,
		Message:   r.URL.Query().Get("msg"),
	})
}

func (h *AdminHandler) RefillAll(w http.ResponseWriter, r *http.Request) {
	n, err := h.admin.RefillAllHearts(r.Context())
	if err != nil {
		http.Error(w, "refill failed", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/admin?msg=Refilled+hearts+for+%d+user(s)", n), http.StatusSeeOther)
}

func (h *AdminHandler) RefillOne(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.admin.RefillUserHearts(r.Context(), id); err != nil {
		http.Error(w, "refill failed", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin?msg=Hearts+refilled", http.StatusSeeOther)
}
