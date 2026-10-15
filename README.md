# 41 — Opportunity follow-up

## Goal

Prepare a next-step message from supplied opportunity facts and last-contact text without reading a CRM or sending an email.

## Run in the host module

This repository is a source-only snapshot and is not runnable by itself: it has no `go.mod` file and does not include the `internal/adkhelper` package imported by `main.go`. Go's `internal` package rule also requires this case to live within the `github.com/GhandyP/adk-go-enterprise-cookbook` module.

To run it, use this case within a checkout of that host module, then run from the case directory:

```bash
export GOOGLE_API_KEY="your-key"
go run .
```

`internal/adkhelper.Run` uses ADK Go v2.2.0, Gemini, and `gemini-flash-latest`; set `ADK_MODEL` to override the model.

Sample prompt:

```text
Prepare a follow-up for this stalled opportunity using the last contact.
```

## What exists now

- `main.go (exists)` sends opportunity facts and last-contact text to one focused Gemini ADK agent.
- The response is free-form; opportunity ID, stage, contact, next action, claims, and timing are not typed.
- This is intentionally text-only: it does not connect to CRM records, activity history, calendar, or an email sender.

## Missing elements

- **Contract:** Define `OpportunityInput{id, stage, facts[], lastContact, goal}` -> `FollowUpPlan{recipient, message, nextAction, suggestedAt, evidence[]}`.
- **ADK layer:** Add an opportunity session and draft schema; expose read-only `get_opportunity` and keep `send_email`/`update_stage` approval-gated.
- **Domain boundary:** Add CRM, calendar, approved-claims, and email adapters with duplicate/outreach-policy checks.
- **Fixtures and validation:** Cover stalled, closed-lost, opted-out, missing-contact, stale-stage, and multi-contact opportunities.
- **Safety:** Never contact an opted-out person or change pipeline state automatically; require a sales-owner approval and dry-run preview.

## Target structure

```text
41-opportunity-follow-up/
├── main.go (exists)
├── README.md (exists)
├── schema.go (next)
├── workflow.go (next)
├── adapters/crm.go (next)
├── testdata/opportunity.json (next)
└── main_test.go (next)
```

## Implementation order

1. Define opportunity, contact, draft, timing, and evidence schemas.
2. Add read-only CRM lookup and a session keyed by opportunity ID.
3. Implement stage/opt-out/duplicate checks and calendar suggestions.
4. Test stale, closed, missing-contact, and multi-contact cases.
5. Add dry-run email and stage-update tools behind owner approval.

## Done when

- [ ] A draft includes the correct opportunity, recipient, evidence, and next action.
- [ ] Opt-outs, closed opportunities, and stale data stop the workflow.
- [ ] Fixtures prove duplicate and missing-contact behavior.
- [ ] Email or CRM updates require a named sales-owner approval.

## Next improvement

Define opportunity eligibility and a typed `FollowUpPlan` before adding CRM or sender tools.
