package opportunity

import (
	"strings"
)

func AssessOpportunity(input OpportunityInput) (Assessment, error) {
	if err := ValidateOpportunityInput(input); err != nil {
		return Assessment{}, err
	}

	if input.OptedOut {
		return Assessment{Status: StatusBlocked, Reason: ReasonOptedOut}, nil
	}
	if isClosedStage(input.Stage) {
		return Assessment{Status: StatusBlocked, Reason: ReasonClosed}, nil
	}
	if input.Stale {
		return Assessment{Status: StatusBlocked, Reason: ReasonStale}, nil
	}

	questions := make([]string, 0, 5)
	if strings.TrimSpace(string(input.Stage)) == "" {
		questions = append(questions, "Confirm the current opportunity stage.")
	}
	if len(input.Facts) == 0 {
		questions = append(questions, "Provide opportunity facts.")
	}
	if strings.TrimSpace(input.LastContact) == "" {
		questions = append(questions, "Provide the last-contact text.")
	}
	if strings.TrimSpace(input.Goal) == "" {
		questions = append(questions, "Describe the desired outcome.")
	}

	recipient, contactQuestion := resolveRecipient(input)
	if contactQuestion != "" {
		questions = append(questions, contactQuestion)
	}
	if len(questions) > 0 {
		return Assessment{
			Status:    StatusNeedsInput,
			Reason:    ReasonMissingInformation,
			Questions: questions,
			Recipient: recipient,
		}, nil
	}

	return Assessment{
		Status:    StatusReady,
		Reason:    ReasonReady,
		Recipient: recipient,
	}, nil
}

func resolveRecipient(input OpportunityInput) (*Contact, string) {
	if len(input.Contacts) == 0 {
		return nil, "Provide at least one contact."
	}

	selectedID := strings.TrimSpace(input.SelectedContactID)
	if selectedID == "" && len(input.Contacts) == 1 {
		recipient := input.Contacts[0]
		return &recipient, ""
	}
	if selectedID == "" {
		return nil, "Select one contact for this follow-up."
	}

	for _, contact := range input.Contacts {
		if strings.TrimSpace(contact.ID) == selectedID {
			recipient := contact
			return &recipient, ""
		}
	}
	return nil, "Select a contact from the provided list."
}

func isClosedStage(stage Stage) bool {
	switch normalizeStage(stage) {
	case "closed", "closed-won", "closed-lost":
		return true
	default:
		return false
	}
}

func normalizeStage(stage Stage) string {
	normalized := strings.ToLower(strings.TrimSpace(string(stage)))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	return strings.Join(strings.Fields(normalized), "-")
}
