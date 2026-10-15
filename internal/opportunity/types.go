package opportunity

import (
	"errors"
	"time"
)

const LastContactEvidenceID = "last_contact"

type Stage string

type Fact struct {
	ID   string
	Text string
}

type Contact struct {
	ID   string
	Name string
}

type OpportunityInput struct {
	ID                string
	Stage             Stage
	Facts             []Fact
	LastContact       string
	Goal              string
	Contacts          []Contact
	SelectedContactID string
	OptedOut          bool
	Stale             bool
}

type FollowUpPlan struct {
	RecipientID string
	Message     string
	NextAction  string
	SuggestedAt *time.Time
	EvidenceIDs []string
}

type DecisionStatus string

const (
	StatusReady      DecisionStatus = "ready"
	StatusNeedsInput DecisionStatus = "needs_input"
	StatusBlocked    DecisionStatus = "blocked"
)

type DecisionReason string

const (
	ReasonReady              DecisionReason = "ready"
	ReasonOptedOut           DecisionReason = "contact_opted_out"
	ReasonClosed             DecisionReason = "opportunity_closed"
	ReasonStale              DecisionReason = "explicitly_marked_stale"
	ReasonMissingInformation DecisionReason = "missing_information"
)

type Assessment struct {
	Status    DecisionStatus
	Reason    DecisionReason
	Questions []string
	Recipient *Contact
}

var (
	ErrInvalidInput       = errors.New("invalid opportunity input")
	ErrNotEligible        = errors.New("opportunity is not eligible for a follow-up draft")
	ErrAssessmentMismatch = errors.New("assessment does not match the opportunity input")
	ErrRecipientMismatch  = errors.New("plan recipient does not match the selected contact")
	ErrEmptyMessage       = errors.New("plan message is required")
	ErrEmptyNextAction    = errors.New("plan next action is required")
	ErrMissingEvidence    = errors.New("plan evidence is required")
	ErrUnknownEvidence    = errors.New("plan references evidence not supplied in the input")
	ErrDuplicateEvidence  = errors.New("plan contains duplicate evidence references")
)
