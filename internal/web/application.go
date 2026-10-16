package web

import (
	"bytes"
	"embed"
	"errors"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/GhandyP/opportunity-follow-up/internal/agent"
	"github.com/GhandyP/opportunity-follow-up/internal/history"
	"github.com/GhandyP/opportunity-follow-up/internal/opportunity"
)

const maxFormBodyBytes = 1 << 20

//go:embed templates/*.html static/app.css
var applicationAssets embed.FS

type applicationHandler struct {
	service    *agent.Service
	store      *history.Store
	templates  *template.Template
	stylesheet []byte
}

type opportunityForm struct {
	ID          string
	Stage       string
	FactsText   string
	LastContact string
	Goal        string
	ContactID   string
	ContactName string
	OptedOut    bool
	Stale       bool
}

type evidenceReference struct {
	ID   string
	Text string
}

type pageData struct {
	Title        string
	ActiveNav    string
	Form         opportunityForm
	Notice       string
	Record       *history.Record
	Records      []history.Record
	Questions    []string
	BlockMessage string
	Evidence     []evidenceReference
	StatusLabel  string
	StatusClass  string
}

func NewApplicationHandler(service *agent.Service, store *history.Store) (http.Handler, error) {
	if service == nil || store == nil {
		return nil, errors.New("agent service and history store are required")
	}

	templates, err := template.New("application").Funcs(template.FuncMap{
		"formatTime":  formatTime,
		"statusLabel": statusLabel,
		"statusClass": statusClass,
	}).ParseFS(applicationAssets, "templates/*.html")
	if err != nil {
		return nil, errors.New("parse application templates")
	}
	stylesheet, err := applicationAssets.ReadFile("static/app.css")
	if err != nil {
		return nil, errors.New("read application stylesheet")
	}

	app := &applicationHandler{
		service:    service,
		store:      store,
		templates:  templates,
		stylesheet: stylesheet,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", health)
	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("POST /opportunities", app.createOpportunity)
	mux.HandleFunc("GET /history", app.historyList)
	mux.HandleFunc("GET /history/{id}", app.historyDetail)
	mux.HandleFunc("POST /history/{id}/draft", app.updateDraft)
	mux.HandleFunc("GET /static/app.css", app.serveStylesheet)

	return secureHeaders(mux), nil
}

func (h *applicationHandler) home(w http.ResponseWriter, _ *http.Request) {
	h.render(w, "home", http.StatusOK, pageData{
		Title:     "New opportunity",
		ActiveNav: "new",
	})
}

func (h *applicationHandler) createOpportunity(w http.ResponseWriter, r *http.Request) {
	if !localFormOriginAllowed(r) {
		http.Error(w, "Cross-origin form submission rejected.", http.StatusForbidden)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxFormBodyBytes)
	if err := r.ParseForm(); err != nil {
		h.render(w, "home", http.StatusBadRequest, pageData{
			Title:     "New opportunity",
			ActiveNav: "new",
			Notice:    "The form could not be read. Check its values and try again.",
			Form:      formFromValues(r.PostForm),
		})
		return
	}

	form := formFromValues(r.PostForm)
	input, err := form.opportunityInput()
	if err != nil {
		h.render(w, "home", http.StatusBadRequest, pageData{
			Title:     "New opportunity",
			ActiveNav: "new",
			Notice:    err.Error(),
			Form:      form,
		})
		return
	}

	result, prepareErr := h.service.Prepare(r.Context(), input)
	if result.Assessment.Status == "" {
		h.render(w, "home", http.StatusUnprocessableEntity, pageData{
			Title:     "New opportunity",
			ActiveNav: "new",
			Notice:    "The opportunity could not be assessed. Review the supplied fields and try again.",
			Form:      form,
		})
		return
	}

	record, err := h.store.Create(r.Context(), input, result)
	if err != nil {
		h.render(w, "home", http.StatusInternalServerError, pageData{
			Title:     "New opportunity",
			ActiveNav: "new",
			Notice:    "The opportunity could not be saved to local history. No message was sent.",
			Form:      form,
		})
		return
	}

	if prepareErr != nil {
		data := pageForRecord(record)
		data.Notice = "The draft could not be generated. Your opportunity details were saved locally; check the server configuration before trying again."
		h.render(w, "detail", http.StatusServiceUnavailable, data)
		return
	}

	http.Redirect(w, r, "/history/"+strconv.FormatInt(record.ID, 10), http.StatusSeeOther)
}

func (h *applicationHandler) historyList(w http.ResponseWriter, r *http.Request) {
	records, err := h.store.List(r.Context())
	if err != nil {
		h.render(w, "history", http.StatusInternalServerError, pageData{
			Title:     "Local history",
			ActiveNav: "history",
			Notice:    "Local history could not be loaded. Try again later.",
		})
		return
	}
	h.render(w, "history", http.StatusOK, pageData{
		Title:     "Local history",
		ActiveNav: "history",
		Records:   records,
	})
}

func (h *applicationHandler) historyDetail(w http.ResponseWriter, r *http.Request) {
	id, ok := recordID(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	record, err := h.store.Get(r.Context(), id)
	if errors.Is(err, history.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "Local history could not be loaded.", http.StatusInternalServerError)
		return
	}
	h.render(w, "detail", http.StatusOK, pageForRecord(record))
}

func (h *applicationHandler) updateDraft(w http.ResponseWriter, r *http.Request) {
	if !localFormOriginAllowed(r) {
		http.Error(w, "Cross-origin form submission rejected.", http.StatusForbidden)
		return
	}
	id, ok := recordID(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxFormBodyBytes)
	if err := r.ParseForm(); err != nil {
		h.renderDraftError(w, r, id, "The draft form could not be read. Check its values and try again.", http.StatusBadRequest)
		return
	}

	_, err := h.store.UpdateDraft(r.Context(), id, r.FormValue("draft_message"), r.FormValue("draft_next_action"))
	if errors.Is(err, history.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, history.ErrInvalidDraft) {
		h.renderDraftError(w, r, id, "Both the message and next action are required to save a draft.", http.StatusBadRequest)
		return
	}
	if errors.Is(err, history.ErrDraftUnavailable) {
		h.renderDraftError(w, r, id, "This opportunity has no editable draft.", http.StatusConflict)
		return
	}
	if err != nil {
		h.renderDraftError(w, r, id, "The draft could not be saved to local history.", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/history/"+strconv.FormatInt(id, 10), http.StatusSeeOther)
}

func (h *applicationHandler) renderDraftError(w http.ResponseWriter, r *http.Request, id int64, message string, status int) {
	record, err := h.store.Get(r.Context(), id)
	if errors.Is(err, history.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "Local history could not be loaded.", http.StatusInternalServerError)
		return
	}
	data := pageForRecord(record)
	data.Notice = message
	h.render(w, "detail", status, data)
}

func (h *applicationHandler) serveStylesheet(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(h.stylesheet)
}

func (h *applicationHandler) render(w http.ResponseWriter, name string, status int, data pageData) {
	var body bytes.Buffer
	if err := h.templates.ExecuteTemplate(&body, name, data); err != nil {
		http.Error(w, "The page could not be rendered.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(body.Bytes())
}

func (form opportunityForm) opportunityInput() (opportunity.OpportunityInput, error) {
	facts, err := parseFacts(form.FactsText)
	if err != nil {
		return opportunity.OpportunityInput{}, err
	}
	input := opportunity.OpportunityInput{
		ID:          strings.TrimSpace(form.ID),
		Stage:       opportunity.Stage(strings.TrimSpace(form.Stage)),
		Facts:       facts,
		LastContact: strings.TrimSpace(form.LastContact),
		Goal:        strings.TrimSpace(form.Goal),
		OptedOut:    form.OptedOut,
		Stale:       form.Stale,
	}
	contactID := strings.TrimSpace(form.ContactID)
	contactName := strings.TrimSpace(form.ContactName)
	if contactID != "" || contactName != "" {
		input.Contacts = []opportunity.Contact{{ID: contactID, Name: contactName}}
		input.SelectedContactID = contactID
	}
	if err := opportunity.ValidateOpportunityInput(input); err != nil {
		return opportunity.OpportunityInput{}, err
	}
	return input, nil
}

func formFromValues(values url.Values) opportunityForm {
	return opportunityForm{
		ID:          values.Get("opportunity_id"),
		Stage:       values.Get("stage"),
		FactsText:   values.Get("facts"),
		LastContact: values.Get("last_contact"),
		Goal:        values.Get("goal"),
		ContactID:   values.Get("contact_id"),
		ContactName: values.Get("contact_name"),
		OptedOut:    formCheckbox(values.Get("opted_out")),
		Stale:       formCheckbox(values.Get("stale")),
	}
}

func localFormOriginAllowed(r *http.Request) bool {
	originHeader := r.Header.Get("Origin")
	if originHeader == "" {
		return true
	}
	origin, err := url.Parse(originHeader)
	if err != nil || (!strings.EqualFold(origin.Scheme, "http") && !strings.EqualFold(origin.Scheme, "https")) || origin.Host == "" || origin.User != nil || origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" {
		return false
	}
	if !strings.EqualFold(origin.Host, r.Host) {
		return false
	}
	host := origin.Hostname()
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func formCheckbox(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "on", "true", "yes":
		return true
	default:
		return false
	}
}

func parseFacts(value string) ([]opportunity.Fact, error) {
	lines := strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n")
	facts := make([]opportunity.Fact, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		id, text, ok := strings.Cut(line, "|")
		id, text = strings.TrimSpace(id), strings.TrimSpace(text)
		if !ok || id == "" || text == "" {
			return nil, errors.New("Each fact must use the format “ID | fact text,” one fact per line.")
		}
		facts = append(facts, opportunity.Fact{ID: id, Text: text})
	}
	return facts, nil
}

func pageForRecord(record history.Record) pageData {
	data := pageData{
		Title:       "Opportunity review",
		ActiveNav:   "history",
		Record:      &record,
		StatusLabel: statusLabel(record.Assessment.Status),
		StatusClass: statusClass(record.Assessment.Status),
	}
	switch record.Assessment.Status {
	case opportunity.StatusBlocked:
		data.BlockMessage = blockMessage(record.Assessment.Reason)
	case opportunity.StatusNeedsInput:
		data.Questions = record.Assessment.Questions
	}
	if record.Plan != nil {
		data.Evidence = evidenceFor(record.Input, *record.Plan)
	}
	return data
}

func evidenceFor(input opportunity.OpportunityInput, plan opportunity.FollowUpPlan) []evidenceReference {
	items := make([]evidenceReference, 0, len(plan.EvidenceIDs))
	for _, id := range plan.EvidenceIDs {
		if id == opportunity.LastContactEvidenceID {
			items = append(items, evidenceReference{ID: id, Text: input.LastContact})
			continue
		}
		for _, fact := range input.Facts {
			if strings.TrimSpace(fact.ID) == strings.TrimSpace(id) {
				items = append(items, evidenceReference{ID: id, Text: fact.Text})
				break
			}
		}
	}
	return items
}

func blockMessage(reason opportunity.DecisionReason) string {
	switch reason {
	case opportunity.ReasonOptedOut:
		return "The contact has opted out. No draft was generated."
	case opportunity.ReasonClosed:
		return "This opportunity is closed. No draft was generated."
	case opportunity.ReasonStale:
		return "This record was marked stale by the owner. No draft was generated."
	default:
		return "A draft is not available for this opportunity."
	}
}

func recordID(value string) (int64, bool) {
	id, err := strconv.ParseInt(value, 10, 64)
	return id, err == nil && id > 0
}

func formatTime(value time.Time) string {
	return value.UTC().Format("Jan 2, 2006 · 15:04 UTC")
}

func statusLabel(status opportunity.DecisionStatus) string {
	switch status {
	case opportunity.StatusReady:
		return "Ready for owner review"
	case opportunity.StatusNeedsInput:
		return "More information needed"
	case opportunity.StatusBlocked:
		return "Follow-up blocked"
	default:
		return "Not assessed"
	}
}

func statusClass(status opportunity.DecisionStatus) string {
	switch status {
	case opportunity.StatusReady:
		return "ready"
	case opportunity.StatusNeedsInput:
		return "needs-input"
	case opportunity.StatusBlocked:
		return "blocked"
	default:
		return "neutral"
	}
}

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}
