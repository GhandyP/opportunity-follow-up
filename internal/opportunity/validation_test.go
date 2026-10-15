package opportunity

import (
	"errors"
	"testing"
)

func TestValidateFollowUpPlan(t *testing.T) {
	baseInput := validOpportunity()
	baseAssessment, err := AssessOpportunity(baseInput)
	if err != nil {
		t.Fatalf("AssessOpportunity() error = %v", err)
	}
	blockedInput := optedOutOpportunity()
	blockedAssessment, err := AssessOpportunity(blockedInput)
	if err != nil {
		t.Fatalf("AssessOpportunity() error = %v", err)
	}

	tests := []struct {
		name       string
		input      OpportunityInput
		assessment Assessment
		plan       FollowUpPlan
		wantErr    error
	}{
		{
			name:       "accepts evidence from supplied facts and last contact",
			input:      baseInput,
			assessment: baseAssessment,
			plan:       validPlan(),
		},
		{
			name:       "rejects a non-eligible opportunity",
			input:      blockedInput,
			assessment: blockedAssessment,
			plan:       validPlan(),
			wantErr:    ErrNotEligible,
		},
		{
			name:       "rejects a different recipient",
			input:      baseInput,
			assessment: baseAssessment,
			plan:       FollowUpPlan{RecipientID: "contact-2", Message: "Hello", NextAction: "Follow up", EvidenceIDs: []string{"fact-1"}},
			wantErr:    ErrRecipientMismatch,
		},
		{
			name:       "rejects an empty message",
			input:      baseInput,
			assessment: baseAssessment,
			plan:       FollowUpPlan{RecipientID: "contact-1", NextAction: "Follow up", EvidenceIDs: []string{"fact-1"}},
			wantErr:    ErrEmptyMessage,
		},
		{
			name:       "rejects an empty next action",
			input:      baseInput,
			assessment: baseAssessment,
			plan:       FollowUpPlan{RecipientID: "contact-1", Message: "Hello", EvidenceIDs: []string{"fact-1"}},
			wantErr:    ErrEmptyNextAction,
		},
		{
			name:       "rejects a plan without evidence",
			input:      baseInput,
			assessment: baseAssessment,
			plan:       FollowUpPlan{RecipientID: "contact-1", Message: "Hello", NextAction: "Follow up"},
			wantErr:    ErrMissingEvidence,
		},
		{
			name:       "rejects evidence absent from the input",
			input:      baseInput,
			assessment: baseAssessment,
			plan:       FollowUpPlan{RecipientID: "contact-1", Message: "Hello", NextAction: "Follow up", EvidenceIDs: []string{"invented-fact"}},
			wantErr:    ErrUnknownEvidence,
		},
		{
			name:       "rejects duplicate evidence references",
			input:      baseInput,
			assessment: baseAssessment,
			plan:       FollowUpPlan{RecipientID: "contact-1", Message: "Hello", NextAction: "Follow up", EvidenceIDs: []string{"fact-1", "fact-1"}},
			wantErr:    ErrDuplicateEvidence,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFollowUpPlan(tt.input, tt.assessment, tt.plan)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ValidateFollowUpPlan() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateFollowUpPlan() error = %v", err)
			}
		})
	}
}

func TestValidateFollowUpPlanRejectsMismatchedAssessment(t *testing.T) {
	input := validOpportunity()
	assessment, err := AssessOpportunity(input)
	if err != nil {
		t.Fatalf("AssessOpportunity() error = %v", err)
	}
	assessment.Recipient = &Contact{ID: "contact-2", Name: "Unlisted Contact"}

	err = ValidateFollowUpPlan(input, assessment, validPlan())
	if !errors.Is(err, ErrAssessmentMismatch) {
		t.Fatalf("ValidateFollowUpPlan() error = %v, want %v", err, ErrAssessmentMismatch)
	}
}

func validPlan() FollowUpPlan {
	return FollowUpPlan{
		RecipientID: "contact-1",
		Message:     "Hi Alex, is the budget review complete?",
		NextAction:  "Ask for a status update.",
		EvidenceIDs: []string{"fact-1", LastContactEvidenceID},
	}
}
