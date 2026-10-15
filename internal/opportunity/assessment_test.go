package opportunity

import (
	"reflect"
	"testing"
)

func TestAssessOpportunity(t *testing.T) {
	tests := []struct {
		name string
		edit func(*OpportunityInput)
		want Assessment
	}{
		{
			name: "stalled opportunity with one contact is ready",
			want: Assessment{
				Status:    StatusReady,
				Reason:    ReasonReady,
				Recipient: &Contact{ID: "contact-1", Name: "Alex Morgan"},
			},
		},
		{
			name: "opted-out opportunity is blocked before requesting missing information",
			edit: func(input *OpportunityInput) {
				input.OptedOut = true
				input.Facts = nil
				input.LastContact = ""
				input.Goal = ""
				input.Contacts = nil
			},
			want: Assessment{Status: StatusBlocked, Reason: ReasonOptedOut},
		},
		{
			name: "closed-lost opportunity is blocked",
			edit: func(input *OpportunityInput) {
				input.Stage = "closed-lost"
			},
			want: Assessment{Status: StatusBlocked, Reason: ReasonClosed},
		},
		{
			name: "closed-won opportunity is blocked",
			edit: func(input *OpportunityInput) {
				input.Stage = "closed-won"
			},
			want: Assessment{Status: StatusBlocked, Reason: ReasonClosed},
		},
		{
			name: "closed stage is normalized before blocking",
			edit: func(input *OpportunityInput) {
				input.Stage = "Closed Lost"
			},
			want: Assessment{Status: StatusBlocked, Reason: ReasonClosed},
		},
		{
			name: "explicit stale flag blocks without inferring an age threshold",
			edit: func(input *OpportunityInput) {
				input.Stale = true
			},
			want: Assessment{Status: StatusBlocked, Reason: ReasonStale},
		},
		{
			name: "missing stage requests confirmation",
			edit: func(input *OpportunityInput) {
				input.Stage = " "
			},
			want: Assessment{
				Status:    StatusNeedsInput,
				Reason:    ReasonMissingInformation,
				Questions: []string{"Confirm the current opportunity stage."},
				Recipient: &Contact{ID: "contact-1", Name: "Alex Morgan"},
			},
		},
		{
			name: "missing facts last contact and goal requests each item",
			edit: func(input *OpportunityInput) {
				input.Facts = nil
				input.LastContact = " "
				input.Goal = ""
			},
			want: Assessment{
				Status: StatusNeedsInput,
				Reason: ReasonMissingInformation,
				Questions: []string{
					"Provide opportunity facts.",
					"Provide the last-contact text.",
					"Describe the desired outcome.",
				},
				Recipient: &Contact{ID: "contact-1", Name: "Alex Morgan"},
			},
		},
		{
			name: "missing contact requests a recipient",
			edit: func(input *OpportunityInput) {
				input.Contacts = nil
			},
			want: Assessment{
				Status:    StatusNeedsInput,
				Reason:    ReasonMissingInformation,
				Questions: []string{"Provide at least one contact."},
			},
		},
		{
			name: "multiple contacts require explicit selection",
			edit: func(input *OpportunityInput) {
				input.Contacts = append(input.Contacts, Contact{ID: "contact-2", Name: "Sam Rivera"})
			},
			want: Assessment{
				Status:    StatusNeedsInput,
				Reason:    ReasonMissingInformation,
				Questions: []string{"Select one contact for this follow-up."},
			},
		},
		{
			name: "selected contact resolves a multi-contact opportunity",
			edit: func(input *OpportunityInput) {
				input.Contacts = append(input.Contacts, Contact{ID: "contact-2", Name: "Sam Rivera"})
				input.SelectedContactID = "contact-2"
			},
			want: Assessment{
				Status:    StatusReady,
				Reason:    ReasonReady,
				Recipient: &Contact{ID: "contact-2", Name: "Sam Rivera"},
			},
		},
		{
			name: "unknown selected contact asks the user to choose from the list",
			edit: func(input *OpportunityInput) {
				input.SelectedContactID = "not-listed"
			},
			want: Assessment{
				Status:    StatusNeedsInput,
				Reason:    ReasonMissingInformation,
				Questions: []string{"Select a contact from the provided list."},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := validOpportunity()
			if tt.edit != nil {
				tt.edit(&input)
			}

			got, err := AssessOpportunity(input)
			if err != nil {
				t.Fatalf("AssessOpportunity() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AssessOpportunity() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
