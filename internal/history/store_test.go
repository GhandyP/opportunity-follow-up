package history

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/GhandyP/opportunity-follow-up/internal/agent"
	"github.com/GhandyP/opportunity-follow-up/internal/opportunity"
)

func TestStorePersistsAndReopensReviewedDraft(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "history.sqlite")
	input := readyInput()
	assessment, err := opportunity.AssessOpportunity(input)
	if err != nil {
		t.Fatalf("AssessOpportunity() error = %v", err)
	}
	plan := opportunity.FollowUpPlan{
		RecipientID: "contact-1",
		Message:     "Thanks for sharing the launch timeline.",
		NextAction:  "Ask whether a short review next week would help.",
		EvidenceIDs: []string{"fact-1", opportunity.LastContactEvidenceID},
	}

	store := openTestStore(t, path)
	created, err := store.Create(ctx, input, agent.Result{Assessment: assessment, Plan: &plan})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("Create() ID = %d, want positive ID", created.ID)
	}
	if created.DraftMessage != plan.Message || created.DraftNextAction != plan.NextAction {
		t.Fatalf("created draft = (%q, %q), want (%q, %q)", created.DraftMessage, created.DraftNextAction, plan.Message, plan.NextAction)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("created timestamps = (%v, %v), want non-zero", created.CreatedAt, created.UpdatedAt)
	}

	listed, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(listed) != 1 || listed[0].ID != created.ID {
		t.Fatalf("List() = %#v, want one record with ID %d", listed, created.ID)
	}

	updated, err := store.UpdateDraft(ctx, created.ID, "Reviewed message.", "Confirm the owner wants a short call.")
	if err != nil {
		t.Fatalf("UpdateDraft() error = %v", err)
	}
	if updated.DraftMessage != "Reviewed message." || updated.DraftNextAction != "Confirm the owner wants a short call." {
		t.Fatalf("updated draft = (%q, %q), want edited values", updated.DraftMessage, updated.DraftNextAction)
	}

	if err := store.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	reopened := openTestStore(t, path)
	reloaded, err := reopened.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get() after reopen error = %v", err)
	}
	if reloaded.Input.ID != input.ID || reloaded.Assessment.Status != opportunity.StatusReady || reloaded.Plan == nil {
		t.Fatalf("reopened record lost input, assessment, or plan: %#v", reloaded)
	}
	if reloaded.Plan.Message != plan.Message || reloaded.DraftMessage != "Reviewed message." || reloaded.DraftNextAction != "Confirm the owner wants a short call." {
		t.Fatalf("reopened record did not preserve original and reviewed drafts: %#v", reloaded)
	}
}

func TestStoreRejectsInvalidInputWithoutChangingHistory(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, filepath.Join(t.TempDir(), "history.sqlite"))
	input := readyInput()
	assessment, err := opportunity.AssessOpportunity(input)
	if err != nil {
		t.Fatalf("AssessOpportunity() error = %v", err)
	}
	plan := opportunity.FollowUpPlan{
		RecipientID: "contact-1",
		Message:     "Evidence-backed draft.",
		NextAction:  "Review the suggested next step.",
		EvidenceIDs: []string{"fact-1"},
	}
	original, err := store.Create(ctx, input, agent.Result{Assessment: assessment, Plan: &plan})
	if err != nil {
		t.Fatalf("Create(valid) error = %v", err)
	}

	invalid := input
	invalid.Facts = append(append([]opportunity.Fact(nil), input.Facts...), opportunity.Fact{ID: "fact-1", Text: "Duplicate fact."})
	_, err = store.Create(ctx, invalid, agent.Result{Assessment: assessment, Plan: &plan})
	if !errors.Is(err, opportunity.ErrInvalidInput) {
		t.Fatalf("Create(invalid) error = %v, want %v", err, opportunity.ErrInvalidInput)
	}
	if got, want := err.Error(), `invalid opportunity input: duplicate fact id "fact-1"`; got != want {
		t.Fatalf("Create(invalid) error = %q, want %q", got, want)
	}

	listed, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() after rejected input error = %v", err)
	}
	if len(listed) != 1 || listed[0].ID != original.ID {
		t.Fatalf("List() after rejected input = %#v, want only original ID %d", listed, original.ID)
	}
	reloaded, err := store.Get(ctx, original.ID)
	if err != nil {
		t.Fatalf("Get(original) after rejected input error = %v", err)
	}
	if reloaded.Input.ID != original.Input.ID || reloaded.Plan == nil || reloaded.Plan.Message != original.Plan.Message {
		t.Fatalf("rejected input changed original history: %#v", reloaded)
	}
}

func TestStoreRejectsBlankDraftWithoutChangingSavedDraft(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, filepath.Join(t.TempDir(), "history.sqlite"))
	input := readyInput()
	assessment, err := opportunity.AssessOpportunity(input)
	if err != nil {
		t.Fatalf("AssessOpportunity() error = %v", err)
	}
	plan := opportunity.FollowUpPlan{
		RecipientID: "contact-1",
		Message:     "Original draft.",
		NextAction:  "Review the evidence.",
		EvidenceIDs: []string{"fact-1"},
	}
	created, err := store.Create(ctx, input, agent.Result{Assessment: assessment, Plan: &plan})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	_, err = store.UpdateDraft(ctx, created.ID, "  \n", "A non-empty action")
	if !errors.Is(err, ErrInvalidDraft) {
		t.Fatalf("UpdateDraft(blank) error = %v, want %v", err, ErrInvalidDraft)
	}
	if got, want := err.Error(), "draft message and next action are required"; got != want {
		t.Fatalf("UpdateDraft(blank) error = %q, want %q", got, want)
	}

	reloaded, err := store.Get(ctx, created.ID)
	if err != nil {
		t.Fatalf("Get() after rejected edit error = %v", err)
	}
	if reloaded.DraftMessage != plan.Message || reloaded.DraftNextAction != plan.NextAction {
		t.Fatalf("rejected edit changed saved draft to (%q, %q)", reloaded.DraftMessage, reloaded.DraftNextAction)
	}
}

func TestStoreKeepsBlockedInputWithoutGeneratingDraft(t *testing.T) {
	ctx := context.Background()
	store := openTestStore(t, filepath.Join(t.TempDir(), "history.sqlite"))
	input := readyInput()
	input.OptedOut = true
	assessment, err := opportunity.AssessOpportunity(input)
	if err != nil {
		t.Fatalf("AssessOpportunity() error = %v", err)
	}
	created, err := store.Create(ctx, input, agent.Result{Assessment: assessment})
	if err != nil {
		t.Fatalf("Create(blocked) error = %v", err)
	}
	if created.Assessment.Status != opportunity.StatusBlocked || created.Assessment.Reason != opportunity.ReasonOptedOut {
		t.Fatalf("blocked assessment = (%q, %q), want (%q, %q)", created.Assessment.Status, created.Assessment.Reason, opportunity.StatusBlocked, opportunity.ReasonOptedOut)
	}
	if created.Plan != nil || created.DraftMessage != "" || created.DraftNextAction != "" {
		t.Fatalf("blocked record contains a follow-up draft: %#v", created)
	}
}

func readyInput() opportunity.OpportunityInput {
	return opportunity.OpportunityInput{
		ID:                "opp-123",
		Stage:             "discovery",
		Facts:             []opportunity.Fact{{ID: "fact-1", Text: "Launch review is planned for next month."}},
		LastContact:       "The customer asked for a short follow-up.",
		Goal:              "Agree on a useful next step.",
		Contacts:          []opportunity.Contact{{ID: "contact-1", Name: "Morgan Lee"}},
		SelectedContactID: "contact-1",
	}
}

func openTestStore(t *testing.T, path string) *Store {
	t.Helper()
	store, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("Open(%q) error = %v", path, err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}
