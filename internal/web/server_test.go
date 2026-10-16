package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/GhandyP/opportunity-follow-up/internal/agent"
	"github.com/GhandyP/opportunity-follow-up/internal/history"
	"github.com/GhandyP/opportunity-follow-up/internal/opportunity"
)

func TestHealthEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		wantStatus int
		wantAllow  string
		wantJSON   bool
	}{
		{
			name:       "reports healthy status",
			method:     http.MethodGet,
			wantStatus: http.StatusOK,
			wantJSON:   true,
		},
		{
			name:       "rejects unsupported method",
			method:     http.MethodPost,
			wantStatus: http.StatusMethodNotAllowed,
			wantAllow:  http.MethodGet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/healthz", nil)
			rec := httptest.NewRecorder()

			NewHandler().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %q", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if got := rec.Header().Get("Allow"); got != tt.wantAllow {
				t.Errorf("Allow header = %q, want %q", got, tt.wantAllow)
			}
			if tt.wantJSON {
				if got := rec.Header().Get("Content-Type"); got != "application/json" {
					t.Fatalf("Content-Type = %q, want application/json", got)
				}

				var body struct {
					Status string `json:"status"`
				}
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if body.Status != "ok" {
					t.Errorf("status body = %q, want ok", body.Status)
				}
			}
		})
	}
}

func TestApplicationWorkflowStoresAndEditsDraft(t *testing.T) {
	generator := &testGenerator{plan: opportunity.FollowUpPlan{
		RecipientID: "C-01",
		Message:     "Hi Jordan, following up about the two-hub pilot.",
		NextAction:  "Review the pilot scope with Jordan",
		EvidenceIDs: []string{"F-01"},
	}}
	handler := newApplicationTestHandler(t, generator)

	home := performRequest(handler, http.MethodGet, "/", nil)
	if home.Code != http.StatusOK {
		t.Fatalf("home status = %d, want %d", home.Code, http.StatusOK)
	}
	for _, want := range []string{"Opportunity details", "Draft review", "LOCAL WORKSPACE", "NOT SENT", "name=\"facts\""} {
		if !strings.Contains(home.Body.String(), want) {
			t.Errorf("home page missing %q", want)
		}
	}

	created := performRequest(handler, http.MethodPost, "/opportunities", validOpportunityForm())
	if created.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d; body=%q", created.Code, http.StatusSeeOther, created.Body.String())
	}
	if got := created.Header().Get("Location"); got != "/history/1" {
		t.Fatalf("create Location = %q, want /history/1", got)
	}
	if generator.calls != 1 {
		t.Fatalf("generator calls = %d, want 1", generator.calls)
	}

	detail := performRequest(handler, http.MethodGet, "/history/1", nil)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail status = %d, want %d; body=%q", detail.Code, http.StatusOK, detail.Body.String())
	}
	for _, want := range []string{"Hi Jordan, following up about the two-hub pilot.", "F-01", "Not sent", "name=\"draft_message\""} {
		if !strings.Contains(detail.Body.String(), want) {
			t.Errorf("detail page missing %q", want)
		}
	}
	if strings.Contains(detail.Body.String(), ">Send</button>") || strings.Contains(detail.Body.String(), "Send now") {
		t.Fatal("detail page must not expose a send action")
	}

	updated := url.Values{
		"draft_message":     {"Updated draft for Jordan."},
		"draft_next_action": {"Confirm pilot timing with the owner."},
	}
	response := performRequest(handler, http.MethodPost, "/history/1/draft", updated)
	if response.Code != http.StatusSeeOther {
		t.Fatalf("draft update status = %d, want %d; body=%q", response.Code, http.StatusSeeOther, response.Body.String())
	}
	if got := response.Header().Get("Location"); got != "/history/1" {
		t.Fatalf("draft update Location = %q, want /history/1", got)
	}

	detail = performRequest(handler, http.MethodGet, "/history/1", nil)
	if detail.Code != http.StatusOK {
		t.Fatalf("updated detail status = %d, want %d", detail.Code, http.StatusOK)
	}
	for _, want := range []string{"Updated draft for Jordan.", "Confirm pilot timing with the owner."} {
		if !strings.Contains(detail.Body.String(), want) {
			t.Errorf("updated detail page missing %q", want)
		}
	}

	historyPage := performRequest(handler, http.MethodGet, "/history", nil)
	if historyPage.Code != http.StatusOK {
		t.Fatalf("history status = %d, want %d", historyPage.Code, http.StatusOK)
	}
	for _, want := range []string{"FF-204", "Updated draft for Jordan."} {
		if !strings.Contains(historyPage.Body.String(), want) {
			t.Errorf("history page missing %q", want)
		}
	}
}

func TestApplicationRejectsCrossOriginWrites(t *testing.T) {
	t.Run("opportunity creation", func(t *testing.T) {
		generator := &testGenerator{}
		handler := newApplicationTestHandler(t, generator)
		response := performRequestWithOrigin(handler, http.MethodPost, "/opportunities", validOpportunityForm(), "attacker.example:8080", "http://attacker.example:8080")
		if response.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d; body=%q", response.Code, http.StatusForbidden, response.Body.String())
		}
		if got, want := response.Body.String(), "Cross-origin form submission rejected.\n"; got != want {
			t.Fatalf("body = %q, want %q", got, want)
		}
		if generator.calls != 0 {
			t.Fatalf("generator calls = %d, want 0", generator.calls)
		}
		historyPage := performRequest(handler, http.MethodGet, "/history", nil)
		if !strings.Contains(historyPage.Body.String(), "No opportunities in your local history yet.") {
			t.Fatalf("rejected submission changed local history: %q", historyPage.Body.String())
		}
	})

	t.Run("draft editing", func(t *testing.T) {
		generator := &testGenerator{plan: opportunity.FollowUpPlan{
			RecipientID: "C-01",
			Message:     "Original locally saved draft.",
			NextAction:  "Review the pilot scope with Jordan",
			EvidenceIDs: []string{"F-01"},
		}}
		handler := newApplicationTestHandler(t, generator)
		created := performRequest(handler, http.MethodPost, "/opportunities", validOpportunityForm())
		if created.Code != http.StatusSeeOther {
			t.Fatalf("setup create status = %d, want %d", created.Code, http.StatusSeeOther)
		}
		maliciousDraft := url.Values{
			"draft_message":     {"Cross-origin replacement."},
			"draft_next_action": {"Cross-origin next action."},
		}
		response := performRequestWithOrigin(handler, http.MethodPost, "/history/1/draft", maliciousDraft, "attacker.example:8080", "http://attacker.example:8080")
		if response.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d; body=%q", response.Code, http.StatusForbidden, response.Body.String())
		}
		if got, want := response.Body.String(), "Cross-origin form submission rejected.\n"; got != want {
			t.Fatalf("body = %q, want %q", got, want)
		}
		detail := performRequest(handler, http.MethodGet, "/history/1", nil)
		for _, want := range []string{"Original locally saved draft.", "Review the pilot scope with Jordan"} {
			if !strings.Contains(detail.Body.String(), want) {
				t.Errorf("rejected edit lost existing draft content %q", want)
			}
		}
		if strings.Contains(detail.Body.String(), "Cross-origin replacement.") || strings.Contains(detail.Body.String(), "Cross-origin next action.") {
			t.Fatal("rejected edit changed local draft")
		}
		if generator.calls != 1 {
			t.Fatalf("generator calls = %d, want 1 (initial assessment only)", generator.calls)
		}
	})
}

func TestApplicationWorkflowBlocksIneligibleInputsBeforeGeneration(t *testing.T) {
	tests := []struct {
		name   string
		change func(url.Values)
		want   string
	}{
		{
			name: "opted-out contact",
			change: func(values url.Values) {
				values.Set("opted_out", "true")
			},
			want: "The contact has opted out.",
		},
		{
			name: "closed opportunity",
			change: func(values url.Values) {
				values.Set("stage", "closed-won")
			},
			want: "This opportunity is closed.",
		},
		{
			name: "owner-marked stale opportunity",
			change: func(values url.Values) {
				values.Set("stale", "true")
			},
			want: "This record was marked stale by the owner.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := &testGenerator{plan: opportunity.FollowUpPlan{
				RecipientID: "C-01",
				Message:     "Should not be generated",
				NextAction:  "Should not be generated",
				EvidenceIDs: []string{"F-01"},
			}}
			handler := newApplicationTestHandler(t, generator)
			values := validOpportunityForm()
			tt.change(values)

			created := performRequest(handler, http.MethodPost, "/opportunities", values)
			if created.Code != http.StatusSeeOther {
				t.Fatalf("create status = %d, want %d; body=%q", created.Code, http.StatusSeeOther, created.Body.String())
			}
			detail := performRequest(handler, http.MethodGet, created.Header().Get("Location"), nil)
			if detail.Code != http.StatusOK {
				t.Fatalf("detail status = %d, want %d", detail.Code, http.StatusOK)
			}
			if !strings.Contains(detail.Body.String(), tt.want) {
				t.Errorf("detail body missing block reason %q", tt.want)
			}
			if strings.Contains(detail.Body.String(), "Should not be generated") {
				t.Fatal("blocked opportunity exposed a generated draft")
			}
			if generator.calls != 0 {
				t.Fatalf("generator calls = %d, want 0", generator.calls)
			}
		})
	}
}

func TestApplicationWorkflowAsksForMissingInformation(t *testing.T) {
	generator := &testGenerator{plan: opportunity.FollowUpPlan{
		RecipientID: "C-01",
		Message:     "Should not be generated",
		NextAction:  "Should not be generated",
		EvidenceIDs: []string{"F-01"},
	}}
	handler := newApplicationTestHandler(t, generator)
	values := validOpportunityForm()
	values.Set("stage", "")
	values.Set("facts", "")
	values.Set("last_contact", "")
	values.Set("goal", "")
	values.Set("contact_id", "")
	values.Set("contact_name", "")

	created := performRequest(handler, http.MethodPost, "/opportunities", values)
	if created.Code != http.StatusSeeOther {
		t.Fatalf("create status = %d, want %d; body=%q", created.Code, http.StatusSeeOther, created.Body.String())
	}
	detail := performRequest(handler, http.MethodGet, created.Header().Get("Location"), nil)
	if detail.Code != http.StatusOK {
		t.Fatalf("detail status = %d, want %d", detail.Code, http.StatusOK)
	}
	for _, want := range []string{"More information needed", "Confirm the current opportunity stage.", "Provide opportunity facts.", "Provide the last-contact text.", "Describe the desired outcome.", "Provide at least one contact."} {
		if !strings.Contains(detail.Body.String(), want) {
			t.Errorf("detail page missing question %q", want)
		}
	}
	if generator.calls != 0 {
		t.Fatalf("generator calls = %d, want 0", generator.calls)
	}
}

func TestApplicationWorkflowRejectsMalformedFactsWithoutSaving(t *testing.T) {
	generator := &testGenerator{}
	handler := newApplicationTestHandler(t, generator)
	values := validOpportunityForm()
	values.Set("facts", "fact with no separator")

	response := performRequest(handler, http.MethodPost, "/opportunities", values)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("create status = %d, want %d; body=%q", response.Code, http.StatusBadRequest, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), "Each fact must use") {
		t.Errorf("validation response missing fact format guidance: %q", response.Body.String())
	}
	if generator.calls != 0 {
		t.Fatalf("generator calls = %d, want 0", generator.calls)
	}
	historyPage := performRequest(handler, http.MethodGet, "/history", nil)
	if historyPage.Code != http.StatusOK {
		t.Fatalf("history status = %d, want %d", historyPage.Code, http.StatusOK)
	}
	if !strings.Contains(historyPage.Body.String(), "No opportunities in your local history yet.") {
		t.Fatalf("invalid input unexpectedly entered history: %q", historyPage.Body.String())
	}
}

func TestApplicationHealthEndpoint(t *testing.T) {
	handler := newApplicationTestHandler(t, &testGenerator{})
	response := performRequest(handler, http.MethodGet, "/healthz", nil)
	if response.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("health Content-Type = %q, want application/json", got)
	}
}

type testGenerator struct {
	plan  opportunity.FollowUpPlan
	err   error
	calls int
}

func (g *testGenerator) Generate(_ context.Context, _ opportunity.OpportunityInput, _ opportunity.Assessment) (opportunity.FollowUpPlan, error) {
	g.calls++
	return g.plan, g.err
}

func newApplicationTestHandler(t *testing.T, generator *testGenerator) http.Handler {
	t.Helper()
	store, err := history.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("open test history: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close test history: %v", err)
		}
	})
	handler, err := NewApplicationHandler(agent.NewService(generator), store)
	if err != nil {
		t.Fatalf("create application handler: %v", err)
	}
	return handler
}

func validOpportunityForm() url.Values {
	return url.Values{
		"opportunity_id": {"FF-204"},
		"stage":          {"evaluation"},
		"facts":          {"F-01 | Jordan requested a two-hub pilot."},
		"last_contact":   {"Jordan requested a short pilot outline after a site visit."},
		"goal":           {"Confirm fit for a two-hub pilot."},
		"contact_id":     {"C-01"},
		"contact_name":   {"Jordan Lee"},
	}
}

func performRequest(handler http.Handler, method, target string, values url.Values) *httptest.ResponseRecorder {
	return performRequestWithOrigin(handler, method, target, values, "", "")
}

func performRequestWithOrigin(handler http.Handler, method, target string, values url.Values, host, origin string) *httptest.ResponseRecorder {
	var body strings.Reader
	if values != nil {
		body = *strings.NewReader(values.Encode())
	}
	request := httptest.NewRequest(method, target, &body)
	if host != "" {
		request.Host = host
	}
	if origin != "" {
		request.Header.Set("Origin", origin)
	}
	if values != nil {
		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
