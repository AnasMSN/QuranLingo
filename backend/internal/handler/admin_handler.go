package handler

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	adminmw "quranlingo/backend/internal/middleware"
	"quranlingo/backend/internal/models"
	"quranlingo/backend/internal/service"
)

//go:embed templates/admin.html templates/admin_login.html templates/admin_questions.html
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

var adminLoginTemplate = template.Must(
	template.New("admin_login.html").ParseFS(adminTemplateFS, "templates/admin_login.html"),
)

var adminQuestionsTemplate = template.Must(
	template.New("admin_questions.html").Funcs(template.FuncMap{
		"add1": func(i int) int { return i + 1 },
		"optionAt": func(opts []models.ExerciseOption, i int) models.ExerciseOption {
			if i < len(opts) {
				return opts[i]
			}
			return models.ExerciseOption{}
		},
		// dict builds a map from alternating key/value args, letting the
		// {{template "questionCard" ...}} partial take multiple named
		// parameters (Go templates only pass a single pipeline value).
		"dict": func(pairs ...any) map[string]any {
			m := make(map[string]any, len(pairs)/2)
			for i := 0; i+1 < len(pairs); i += 2 {
				if key, ok := pairs[i].(string); ok {
					m[key] = pairs[i+1]
				}
			}
			return m
		},
	}).ParseFS(adminTemplateFS, "templates/admin_questions.html"),
)

// optionSlots is the fixed 4-slot layout the question forms render — a
// question can use 2, 3, or 4 of them (trailing ones left blank/unchecked).
var optionSlots = []int{0, 1, 2, 3}

// AdminHandler serves the operator-only, server-rendered HTML admin panel.
// It deliberately does not use the httpx JSON helpers — this surface speaks
// HTML, not the mobile app's JSON API.
type AdminHandler struct {
	admin         *service.AdminService
	username      string
	passwordHash  string
	sessionSecret string
	secureCookie  bool
}

func NewAdminHandler(admin *service.AdminService, username, passwordHash, sessionSecret string, secureCookie bool) *AdminHandler {
	return &AdminHandler{
		admin:         admin,
		username:      username,
		passwordHash:  passwordHash,
		sessionSecret: sessionSecret,
		secureCookie:  secureCookie,
	}
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

type adminLoginPageData struct {
	Error string
}

// LoginPage renders the dedicated admin sign-in form. This replaces the
// browser-native HTTP Basic Auth prompt with a real page the operator can
// bookmark, style, and (later) extend with things like a second factor.
func (h *AdminHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// Already signed in? Skip straight to the dashboard.
	if cookie, err := r.Cookie(adminmw.AdminSessionCookie); err == nil && cookie.Value != "" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = adminLoginTemplate.Execute(w, adminLoginPageData{
		Error: r.URL.Query().Get("error"),
	})
}

// Login validates the submitted credentials against the single operator
// credential from config (ADMIN_USERNAME/ADMIN_PASSWORD_HASH) — there is no
// registration flow for admins, so this is the only account that can ever
// pass this check — and issues a signed session cookie on success.
func (h *AdminHandler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/login?error=Invalid+form+submission", http.StatusSeeOther)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if !adminmw.VerifyAdminCredentials(h.username, h.passwordHash, username, password) {
		http.Redirect(w, r, "/admin/login?error=Invalid+username+or+password", http.StatusSeeOther)
		return
	}

	http.SetCookie(w, adminmw.NewAdminSessionCookie(h.sessionSecret, h.username, h.secureCookie))
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// Logout clears the session cookie and sends the operator back to the login page.
func (h *AdminHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, adminmw.ExpiredAdminSessionCookie(h.secureCookie))
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

// --- Question management: manual entry / CSV import, preview, edit, approve ---

type adminQuestionsPageData struct {
	Lessons        []service.LessonOption
	SelectedLesson string
	Drafts         []service.QuestionPreview
	Approved       []service.QuestionPreview
	OptionSlots    []int
	Message        string
	Error          string
}

func (h *AdminHandler) QuestionsPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	lessons, err := h.admin.ListLessonsForPicker(ctx)
	if err != nil {
		http.Error(w, "failed to load lessons", http.StatusInternalServerError)
		return
	}

	lessonID := r.URL.Query().Get("lesson_id")
	if lessonID == "" && len(lessons) > 0 {
		lessonID = lessons[0].ID
	}

	var drafts, approved []service.QuestionPreview
	if lessonID != "" {
		if drafts, err = h.admin.ListDraftQuestions(ctx, lessonID); err != nil {
			http.Error(w, "failed to load draft questions", http.StatusInternalServerError)
			return
		}
		if approved, err = h.admin.ListApprovedQuestions(ctx, lessonID); err != nil {
			http.Error(w, "failed to load approved questions", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = adminQuestionsTemplate.Execute(w, adminQuestionsPageData{
		Lessons:        lessons,
		SelectedLesson: lessonID,
		Drafts:         drafts,
		Approved:       approved,
		OptionSlots:    optionSlots,
		Message:        r.URL.Query().Get("msg"),
		Error:          r.URL.Query().Get("error"),
	})
}

// parseQuestionOptionsForm reads option_1..option_4 and the correct_options
// checkboxes (1-based, matching the option's ordinal, not array index) from
// a submitted question form. Blank option fields are dropped, so a 2, 3, or
// 4-option question can all be submitted from the same fixed-size form.
// More than one option can be checked correct — Arabic words often have
// several valid English glosses.
func parseQuestionOptionsForm(r *http.Request) []service.QuestionOption {
	correctSet := map[string]bool{}
	for _, v := range r.Form["correct_options"] {
		correctSet[v] = true
	}

	var options []service.QuestionOption
	for n := 1; n <= 4; n++ {
		text := strings.TrimSpace(r.FormValue(fmt.Sprintf("option_%d", n)))
		if text == "" {
			continue
		}
		options = append(options, service.QuestionOption{Text: text, Correct: correctSet[strconv.Itoa(n)]})
	}
	return options
}

func questionsRedirectURL(lessonID string, params url.Values) string {
	params.Set("lesson_id", lessonID)
	return "/admin/questions?" + params.Encode()
}

// CreateQuestion adds one manually-entered question to a lesson as a draft —
// it never touches the live app until an operator explicitly approves it.
func (h *AdminHandler) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/questions?error=Invalid+form+submission", http.StatusSeeOther)
		return
	}
	lessonID := r.FormValue("lesson_id")
	options := parseQuestionOptionsForm(r)

	if _, err := h.admin.CreateDraftQuestion(r.Context(), lessonID, r.FormValue("prompt"), r.FormValue("arabic_text"), r.FormValue("audio_url"), options); err != nil {
		http.Redirect(w, r, questionsRedirectURL(lessonID, url.Values{"error": {err.Error()}}), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, questionsRedirectURL(lessonID, url.Values{"msg": {"Question added as a draft — preview it below before approving"}}), http.StatusSeeOther)
}

// ImportQuestions bulk-creates draft questions from an uploaded CSV file.
// Arabic text (including harakat) round-trips fine since the CSV is read as
// plain UTF-8 — no special encoding handling needed beyond stripping a BOM.
func (h *AdminHandler) ImportQuestions(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Redirect(w, r, "/admin/questions?error=Invalid+file+upload", http.StatusSeeOther)
		return
	}
	lessonID := r.FormValue("lesson_id")

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Redirect(w, r, questionsRedirectURL(lessonID, url.Values{"error": {"Choose a CSV file to import"}}), http.StatusSeeOther)
		return
	}
	defer file.Close()

	result, err := h.admin.ImportQuestionsCSV(r.Context(), lessonID, file)
	if err != nil {
		http.Redirect(w, r, questionsRedirectURL(lessonID, url.Values{"error": {err.Error()}}), http.StatusSeeOther)
		return
	}

	params := url.Values{"msg": {fmt.Sprintf("Imported %d question(s) as drafts", result.Created)}}
	if len(result.Errors) > 0 {
		errMsg := fmt.Sprintf("%d row(s) skipped: %s", len(result.Errors), strings.Join(result.Errors, "; "))
		if len(errMsg) > 800 {
			errMsg = errMsg[:800] + "…"
		}
		params.Set("error", errMsg)
	}
	http.Redirect(w, r, questionsRedirectURL(lessonID, params), http.StatusSeeOther)
}

// UpdateQuestion overwrites a question's fields — used both to fix a draft
// spotted in the preview before approving it, and to edit an already-live
// question directly from the "existing questions" view.
func (h *AdminHandler) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/questions?error=Invalid+form+submission", http.StatusSeeOther)
		return
	}
	lessonID := r.FormValue("lesson_id")
	options := parseQuestionOptionsForm(r)

	if err := h.admin.UpdateQuestion(r.Context(), id, r.FormValue("prompt"), r.FormValue("arabic_text"), r.FormValue("audio_url"), options); err != nil {
		http.Redirect(w, r, questionsRedirectURL(lessonID, url.Values{"error": {err.Error()}}), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, questionsRedirectURL(lessonID, url.Values{"msg": {"Question updated"}}), http.StatusSeeOther)
}

// ApproveQuestion publishes a draft so it becomes visible to the mobile app.
func (h *AdminHandler) ApproveQuestion(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	lessonID := r.FormValue("lesson_id")
	if err := h.admin.ApproveQuestion(r.Context(), id); err != nil {
		http.Redirect(w, r, questionsRedirectURL(lessonID, url.Values{"error": {err.Error()}}), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, questionsRedirectURL(lessonID, url.Values{"msg": {"Question approved — now live in the app"}}), http.StatusSeeOther)
}

// DeleteQuestion rejects/removes a draft question that shouldn't be published.
func (h *AdminHandler) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	lessonID := r.FormValue("lesson_id")
	if err := h.admin.DeleteDraftQuestion(r.Context(), id); err != nil {
		http.Redirect(w, r, questionsRedirectURL(lessonID, url.Values{"error": {err.Error()}}), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, questionsRedirectURL(lessonID, url.Values{"msg": {"Draft question deleted"}}), http.StatusSeeOther)
}
