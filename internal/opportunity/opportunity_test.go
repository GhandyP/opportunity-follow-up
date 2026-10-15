package opportunity

import (
	"errors"
	"testing"
)

func TestValidateOpportunityInput(t *testing.T) {
	tests := []struct {
		name    string
		edit    func(*OpportunityInput)
		wantErr bool
	}{
		{
			name: "valid input",
		},
		{
			name: "missing fact id is invalid",
			edit: func(input *OpportunityInput) {
				input.Facts[0].ID = " "
			},
			wantErr: true,
		},
		{
			name: "missing fact text is invalid",
			edit: func(input *OpportunityInput) {
				input.Facts[0].Text = " "
			},
			wantErr: true,
		},
		{
			name: "duplicate fact ids are invalid",
			edit: func(input *OpportunityInput) {
				input.Facts = append(input.Facts, Fact{ID: "fact-1", Text: "Another fact."})
			},
			wantErr: true,
		},
		{
			name: "last-contact evidence id is reserved",
			edit: func(input *OpportunityInput) {
				input.Facts[0].ID = LastContactEvidenceID
			},
			wantErr: true,
		},
		{
			name: "contact without id is invalid",
			edit: func(input *OpportunityInput) {
				input.Contacts[0].ID = " "
			},
			wantErr: true,
		},
		{
			name: "contact without name is invalid",
			edit: func(input *OpportunityInput) {
				input.Contacts[0].Name = " "
			},
			wantErr: true,
		},
		{
			name: "duplicate contact ids are invalid",
			edit: func(input *OpportunityInput) {
				input.Contacts = append(input.Contacts, Contact{ID: "contact-1", Name: "Sam Rivera"})
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validOpportunity()
			if tt.edit != nil {
				tt.edit(&input)
			}

			err := ValidateOpportunityInput(input)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidInput) {
					t.Fatalf("ValidateOpportunityInput() error = %v, want %v", err, ErrInvalidInput)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateOpportunityInput() error = %v", err)
			}
		})
	}
}

func validOpportunity() OpportunityInput {
	return OpportunityInput{
		ID:          "opp-41",
		Stage:       "stalled",
		Facts:       []Fact{{ID: "fact-1", Text: "Budget review is next week."}},
		LastContact: "The prospect asked us to follow up after the budget review.",
		Goal:        "Ask whether the review is complete.",
		Contacts:    []Contact{{ID: "contact-1", Name: "Alex Morgan"}},
	}
}

func optedOutOpportunity() OpportunityInput {
	input := validOpportunity()
	input.OptedOut = true
	return input
}
