package httpx

import (
	"context"
	"embed"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/yplog/ticktockbox/internal/jobs"
)

type AdminHandlers struct {
	Repo        *jobs.Repo
	Scheduler   *jobs.Scheduler
	TemplatesFS embed.FS
	Assets      embed.FS
	Validate    *validator.Validate
}

type createJobForm struct {
	Title               string `validate:"required,min=3"`
	TZ                  string `validate:"required"`
	RunAt               string `validate:"required"`
	RemindBeforeMinutes int    `validate:"min=0,max=10080"`
}

func (a *AdminHandlers) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := r.URL.Query().Get("status")
	if status == "" {
		status = "pending"
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 25
	}

	filter := jobs.JobFilter{
		Status: status,
		Page:   page,
		Limit:  limit,
	}

	jobPage, err := a.Repo.GetJobsPaginated(ctx, filter)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	type row struct {
		jobs.Job
		RunAtLocal string
		DueAtLocal string
	}

	var rows []row
	for _, j := range jobPage.Jobs {
		loc, _ := time.LoadLocation(j.TZ)
		rows = append(rows, row{
			Job:        j,
			RunAtLocal: j.RunAtUTC.In(loc).Format("2006-01-02 15:04:05"),
			DueAtLocal: j.DueAtUTC.In(loc).Format("2006-01-02 15:04:05"),
		})
	}

	data := map[string]any{
		"Rows":       rows,
		"Page":       jobPage,
		"Filter":     filter,
		"StatusList": []string{"all", "pending", "enqueued", "completed", "cancelled"},
	}

	tmpl := template.New("index").Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return string(s[0]-32) + s[1:]
		},
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
	})

	tmpl = template.Must(tmpl.ParseFS(a.TemplatesFS, "layout.tmpl", "index.tmpl"))
	_ = tmpl.ExecuteTemplate(w, "index", data)
}

func (a *AdminHandlers) NewForm(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFS(a.TemplatesFS, "layout.tmpl", "new.tmpl"))
	_ = tmpl.ExecuteTemplate(w, "new", nil)
}

func parseTimeInTZ(s, tz string) (time.Time, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.Time{}, err
	}

	// Önce RFC3339 formatını dene
	if t, err := time.ParseInLocation(time.RFC3339, s, loc); err == nil {
		return t, nil
	}

	// datetime-local input formatını dene (saniye ile)
	if t, err := time.ParseInLocation("2006-01-02T15:04:05", s, loc); err == nil {
		return t, nil
	}

	// datetime-local input formatını dene (saniye olmadan)
	if t, err := time.ParseInLocation("2006-01-02T15:04", s, loc); err == nil {
		return t, nil
	}

	// Eski format (fallback)
	layout := "2006-01-02 15:04"
	return time.ParseInLocation(layout, s, loc)
}

func (a *AdminHandlers) CreateJob(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	mins, _ := strconv.Atoi(r.PostForm.Get("remind_before_minutes"))

	form := createJobForm{
		Title:               r.PostForm.Get("title"),
		TZ:                  r.PostForm.Get("tz"),
		RunAt:               r.PostForm.Get("run_at"),
		RemindBeforeMinutes: mins,
	}

	if err := a.Validate.Struct(form); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	runLocal, err := parseTimeInTZ(form.RunAt, form.TZ)
	if err != nil {
		http.Error(w, "invalid time or timezone: "+err.Error(), 400)
		return
	}

	runUTC := runLocal.UTC()
	dueUTC := runUTC.Add(-time.Duration(form.RemindBeforeMinutes) * time.Minute)

	j := jobs.Job{
		Title:               form.Title,
		TZ:                  form.TZ,
		RunAtUTC:            runUTC,
		DueAtUTC:            dueUTC,
		RemindBeforeMinutes: form.RemindBeforeMinutes,
	}

	ctx := context.Background()

	id, err := a.Repo.Insert(ctx, &j)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	j.ID = id

	a.Scheduler.ScheduleNew(j)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *AdminHandlers) CancelJob(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	if err := a.Repo.Cancel(r.Context(), id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
