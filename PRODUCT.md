# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

Go templates + HTML/CSS

## Users

Sales owners who prepare and review follow-up messages from manually supplied opportunity information. Other user roles and team or workspace structure are unspecified.

## Product Purpose

Prepare an evidence-grounded next-step message from supplied opportunity facts and last-contact text. The product supports human review and editing; it does not send messages or read CRM records.

## Positioning

The specified mechanism combines deterministic eligibility checks with an agent-generated, evidence-referenced draft. It asks for missing information and keeps follow-up subject to human review rather than taking outreach or pipeline actions autonomously.

## Operating Context

The first version runs locally in a browser and stores opportunity and editable-draft history in local SQLite. The repository is public, but the application has no public deployment target. Gemini credentials are supplied server-side.

## Capabilities and Constraints

- Users supply the opportunity, stage, facts, last-contact text, goal, and contact information.
- Missing information produces questions; eligible inputs may produce a structured draft with evidence references.
- Opt-outs, closed opportunities, and data explicitly marked stale block draft generation. The app does not infer staleness from age.
- Drafts are for sales-owner review and editing. There is no automatic email sending, CRM read or update, calendar integration, pipeline change, or multiuser support in v1.
- Keep Gemini credentials out of Git and logs. Use fictional data in examples.
- Local record retention and deletion behavior beyond SQLite persistence are unspecified; do not invent automatic expiry or deletion UX.

## Brand Commitments

The existing project name is “Opportunity follow-up.” No logo, visual identity, or additional brand assets are specified. Preserve the factual tone of the existing README.

## Evidence on Hand

- `README.md` defines the original text-only opportunity follow-up goal and safety constraints.
- `internal/opportunity/` contains typed inputs, eligibility rules, and plan/evidence validation.
- No real customer records, testimonials, benchmarks, or production case studies are provided. Examples must remain fictional.

## Product Principles

1. Enforce eligibility before invoking the model.
2. Ground drafts only in supplied information and ask when required information is missing.
3. Keep a human in control; never send messages or mutate CRM or pipeline state.
4. Keep history local and credentials server-side.
