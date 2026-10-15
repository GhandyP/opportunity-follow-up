package agent

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/GhandyP/opportunity-follow-up/internal/opportunity"
)

func TestPrepareDoesNotGenerateForNonReadyInput(t *testing.T) {
	tests := []struct {
		name       string
		edit       func(*opportunity.OpportunityInput)
		wantStatus opportunity.DecisionStatus
		wantReason opportunity.DecisionReason
		wantQs     []string
	}{
		{
			name: "opted out",
			edit: func(input *opportunity.OpportunityInput) {
				input.OptedOut = true
			},
			wantStatus: opportunity.StatusBlocked,
			wantReason: opportunity.ReasonOptedOut,
		},
		{
			name: "closed opportunity",
			edit: func(input *opportunity.OpportunityInput) {
				input.Stage = "closed-lost"
			},
			wantStatus: opportunity.StatusBlocked,
			wantReason: opportunity.ReasonClosed,
		},
		{
			name: "explicitly stale opportunity",
			edit: func(input *opportunity.OpportunityInput) {
				input.Stale = true
			},
			wantStatus: opportunity.StatusBlocked,
			wantReason: opportunity.ReasonStale,
		},
		{
			name: "missing information",
			edit: func(input *opportunity.OpportunityInput) {
				input.LastContact = " "
			},
			wantStatus: opportunity.StatusNeedsInput,
			wantReason: opportunity.ReasonMissingInformation,
			wantQs:     []string{"Provide the last-contact text."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validInput()
			if tt.edit != nil {
				tt.edit(&input)
			}
			generator := &fakeGenerator{plan: validPlan()}
			service := NewService(generator)

			got, err := service.Prepare(context.Background(), input)
			if err != nil {
				t.Fatalf("Prepare() error = %v", err)
			}
			if got.Assessment.Status != tt.wantStatus || got.Assessment.Reason != tt.wantReason {
				t.Fatalf("assessment = %#v, want status %q and reason %q", got.Assessment, tt.wantStatus, tt.wantReason)
			}
			if !reflect.DeepEqual(got.Assessment.Questions, tt.wantQs) {
				t.Errorf("questions = %#v, want %#v", got.Assessment.Questions, tt.wantQs)
			}
			if got.Plan != nil {
				t.Errorf("plan = %#v, want nil", got.Plan)
			}
			if generator.calls != 0 {
				t.Errorf("generator calls = %d, want 0", generator.calls)
			}
		})
	}
}

func TestPrepareValidatesGeneratedPlan(t *testing.T) {
	input := validInput()
	generator := &fakeGenerator{plan: validPlan()}
	service := NewService(generator)

	got, err := service.Prepare(context.Background(), input)
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	if got.Assessment.Status != opportunity.StatusReady {
		t.Fatalf("assessment status = %q, want %q", got.Assessment.Status, opportunity.StatusReady)
	}
	if got.Plan == nil || !reflect.DeepEqual(*got.Plan, generator.plan) {
		t.Fatalf("plan = %#v, want %#v", got.Plan, generator.plan)
	}
	if generator.calls != 1 {
		t.Fatalf("generator calls = %d, want 1", generator.calls)
	}
	if !reflect.DeepEqual(generator.input, input) {
		t.Errorf("generator input = %#v, want %#v", generator.input, input)
	}
	if generator.assessment.Status != opportunity.StatusReady || generator.assessment.Recipient == nil {
		t.Errorf("generator assessment = %#v, want ready with a recipient", generator.assessment)
	}
}

func TestPrepareRejectsUnsupportedEvidenceFromGenerator(t *testing.T) {
	generator := &fakeGenerator{plan: validPlan()}
	generator.plan.EvidenceIDs = []string{"not-in-input"}
	service := NewService(generator)

	got, err := service.Prepare(context.Background(), validInput())
	if !errors.Is(err, opportunity.ErrUnknownEvidence) {
		t.Fatalf("Prepare() error = %v, want %v", err, opportunity.ErrUnknownEvidence)
	}
	if got.Plan != nil {
		t.Errorf("plan = %#v after invalid evidence, want nil", got.Plan)
	}
	if generator.calls != 1 {
		t.Errorf("generator calls = %d, want 1", generator.calls)
	}
}

func TestPreparePropagatesGeneratorError(t *testing.T) {
	wantErr := errors.New("model unavailable")
	generator := &fakeGenerator{err: wantErr}
	service := NewService(generator)

	got, err := service.Prepare(context.Background(), validInput())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Prepare() error = %v, want %v", err, wantErr)
	}
	if got.Assessment.Status != opportunity.StatusReady {
		t.Errorf("assessment status = %q, want %q", got.Assessment.Status, opportunity.StatusReady)
	}
	if got.Plan != nil {
		t.Errorf("plan = %#v after generator error, want nil", got.Plan)
	}
}

func TestPrepareReportsMissingGeneratorOnlyForReadyInput(t *testing.T) {
	service := NewService(nil)

	blocked := validInput()
	blocked.OptedOut = true
	got, err := service.Prepare(context.Background(), blocked)
	if err != nil {
		t.Fatalf("Prepare() for blocked input error = %v", err)
	}
	if got.Assessment.Status != opportunity.StatusBlocked || got.Plan != nil {
		t.Fatalf("blocked result = %#v, want blocked with no plan", got)
	}

	got, err = service.Prepare(context.Background(), validInput())
	if !errors.Is(err, ErrGeneratorUnavailable) {
		t.Fatalf("Prepare() for ready input error = %v, want %v", err, ErrGeneratorUnavailable)
	}
	if got.Assessment.Status != opportunity.StatusReady || got.Plan != nil {
		t.Fatalf("ready result = %#v, want ready with no plan", got)
	}
}

func TestPrepareRejectsInvalidInputBeforeGenerating(t *testing.T) {
	input := validInput()
	input.Facts[0].ID = " "
	generator := &fakeGenerator{plan: validPlan()}
	service := NewService(generator)

	got, err := service.Prepare(context.Background(), input)
	if !errors.Is(err, opportunity.ErrInvalidInput) {
		t.Fatalf("Prepare() error = %v, want %v", err, opportunity.ErrInvalidInput)
	}
	if got.Plan != nil {
		t.Errorf("plan = %#v for invalid input, want nil", got.Plan)
	}
	if generator.calls != 0 {
		t.Errorf("generator calls = %d, want 0", generator.calls)
	}
}

type fakeGenerator struct {
	plan       opportunity.FollowUpPlan
	err        error
	calls      int
	input      opportunity.OpportunityInput
	assessment opportunity.Assessment
}

func (f *fakeGenerator) Generate(_ context.Context, input opportunity.OpportunityInput, assessment opportunity.Assessment) (opportunity.FollowUpPlan, error) {
	f.calls++
	f.input = input
	f.assessment = assessment
	return f.plan, f.err
}

func validInput() opportunity.OpportunityInput {
	return opportunity.OpportunityInput{
		ID:          "opp-1",
		Stage:       "stalled",
		Facts:       []opportunity.Fact{{ID: "fact-1", Text: "Budget review is next week."}},
		LastContact: "The prospect asked us to follow up after the budget review.",
		Goal:        "Ask whether the review is complete.",
		Contacts:    []opportunity.Contact{{ID: "contact-1", Name: "Alex Morgan"}},
	}
}

func validPlan() opportunity.FollowUpPlan {
	return opportunity.FollowUpPlan{
		RecipientID: "contact-1",
		Message:     "Hi Alex, is the budget review complete?",
		NextAction:  "Ask for a status update.",
		EvidenceIDs: []string{"fact-1", opportunity.LastContactEvidenceID},
	}
}
