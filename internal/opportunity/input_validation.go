package opportunity

import (
	"fmt"
	"strings"
)

func ValidateOpportunityInput(input OpportunityInput) error {
	factIDs := make(map[string]struct{}, len(input.Facts))
	for i, fact := range input.Facts {
		id := strings.TrimSpace(fact.ID)
		if id == "" || strings.TrimSpace(fact.Text) == "" {
			return fmt.Errorf("%w: fact %d requires an id and text", ErrInvalidInput, i+1)
		}
		if id == LastContactEvidenceID {
			return fmt.Errorf("%w: fact id %q is reserved", ErrInvalidInput, id)
		}
		if _, exists := factIDs[id]; exists {
			return fmt.Errorf("%w: duplicate fact id %q", ErrInvalidInput, id)
		}
		factIDs[id] = struct{}{}
	}

	contactIDs := make(map[string]struct{}, len(input.Contacts))
	for i, contact := range input.Contacts {
		id := strings.TrimSpace(contact.ID)
		if id == "" || strings.TrimSpace(contact.Name) == "" {
			return fmt.Errorf("%w: contact %d requires an id and name", ErrInvalidInput, i+1)
		}
		if _, exists := contactIDs[id]; exists {
			return fmt.Errorf("%w: duplicate contact id %q", ErrInvalidInput, id)
		}
		contactIDs[id] = struct{}{}
	}

	return nil
}
