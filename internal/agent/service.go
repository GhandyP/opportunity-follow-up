package agent

import (
	"context"
	"errors"

	"github.com/GhandyP/opportunity-follow-up/internal/opportunity"
)

var ErrGeneratorUnavailable = errors.New("follow-up plan generator is unavailable")

type Generator interface {
	Generate(context.Context, opportunity.OpportunityInput, opportunity.Assessment) (opportunity.FollowUpPlan, error)
}

type Result struct {
	Assessment opportunity.Assessment
	Plan       *opportunity.FollowUpPlan
}

type Service struct {
	generator Generator
}

func NewService(generator Generator) *Service {
	return &Service{generator: generator}
}

func (s *Service) Prepare(ctx context.Context, input opportunity.OpportunityInput) (Result, error) {
	assessment, err := opportunity.AssessOpportunity(input)
	if err != nil {
		return Result{}, err
	}
	result := Result{Assessment: assessment}
	if assessment.Status != opportunity.StatusReady {
		return result, nil
	}
	if s == nil || s.generator == nil {
		return result, ErrGeneratorUnavailable
	}

	plan, err := s.generator.Generate(ctx, input, assessment)
	if err != nil {
		return result, err
	}
	if err := opportunity.ValidateFollowUpPlan(input, assessment, plan); err != nil {
		return result, err
	}

	result.Plan = &plan
	return result, nil
}
