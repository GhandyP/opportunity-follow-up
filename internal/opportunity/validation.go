package opportunity

import "strings"

func ValidateFollowUpPlan(input OpportunityInput, assessment Assessment, plan FollowUpPlan) error {
	if err := ValidateOpportunityInput(input); err != nil {
		return err
	}
	current, err := AssessOpportunity(input)
	if err != nil {
		return err
	}
	if current.Status != StatusReady || assessment.Status != StatusReady {
		return ErrNotEligible
	}
	if current.Recipient == nil || assessment.Recipient == nil ||
		current.Reason != assessment.Reason ||
		strings.TrimSpace(current.Recipient.ID) != strings.TrimSpace(assessment.Recipient.ID) {
		return ErrAssessmentMismatch
	}
	if strings.TrimSpace(plan.RecipientID) != strings.TrimSpace(current.Recipient.ID) {
		return ErrRecipientMismatch
	}
	if strings.TrimSpace(plan.Message) == "" {
		return ErrEmptyMessage
	}
	if strings.TrimSpace(plan.NextAction) == "" {
		return ErrEmptyNextAction
	}
	if len(plan.EvidenceIDs) == 0 {
		return ErrMissingEvidence
	}

	validEvidence := make(map[string]struct{}, len(input.Facts)+1)
	for _, fact := range input.Facts {
		validEvidence[strings.TrimSpace(fact.ID)] = struct{}{}
	}
	if strings.TrimSpace(input.LastContact) != "" {
		validEvidence[LastContactEvidenceID] = struct{}{}
	}

	seen := make(map[string]struct{}, len(plan.EvidenceIDs))
	for _, evidenceID := range plan.EvidenceIDs {
		id := strings.TrimSpace(evidenceID)
		if _, exists := seen[id]; exists {
			return ErrDuplicateEvidence
		}
		seen[id] = struct{}{}
		if _, exists := validEvidence[id]; !exists {
			return ErrUnknownEvidence
		}
	}

	return nil
}
