# Opportunity Follow-up Web App — ODD Feature Plan
Objective: Rebuild the public prototype as a standalone Go web app using Google ADK for safe, evidence-grounded opportunity follow-up guidance.
Branch: `feature/opportunity-follow-up-web-app`; delivery strategy: `ask-on-risk` (resolved); chain strategy: `feature-branch-chain`; forecast: ~650 authored changed lines; no push, PR, or deployment authorized.
Checks: `go test ./...`; browser smoke test for the primary workflow; live Gemini checks only with explicitly supplied credentials.

## Specs

### S1 — Build the application
> ok entonces vamos a construir desde cero este proyecto para que sea una app web con go y google adk y ademas crea el
> ▎ plan con odd y me avisas cuando tengas el plan listo

### S2 — Make the agent more useful
> ok piensa como un agente puede ser mas util y usa la informacion que tienes

### S3 — Keep the repository public
> el repo es publico

### S4 — Preserve the existing project intent and safety constraints
From `README.md`, verbatim:
> Prepare a next-step message from supplied opportunity facts and last-contact text without reading a CRM or sending an email.

> This is intentionally text-only: it does not connect to CRM records, activity history, calendar, or an email sender.

> `OpportunityInput{id, stage, facts[], lastContact, goal}` -> `FollowUpPlan{recipient, message, nextAction, suggestedAt, evidence[]}`.

> Opt-outs, closed opportunities, and stale data stop the workflow.

> Never contact an opted-out person or change pipeline state automatically; require a sales-owner approval and dry-run preview.

### Planning assumptions (not additional user requirements)
- Start as a local-only, server-rendered Go app; public source code does not imply public hosting.
- Use SQLite for local opportunity/draft history and synthetic demo fixtures; do not connect to a real CRM, calendar, or email sender in this scope.
- Keep Gemini credentials in server environment configuration; never commit or log them.
- Deterministic application rules enforce ineligibility before the model runs. The ADK agent asks for missing facts and drafts a typed plan; it does not decide opt-out/closed gates or perform external side effects.
- Confirm these defaults if implementation reveals a consequential product decision; no production data or deployment is assumed.

## Tasks

- T1 | S1,S3 | inline | in_progress | Create a standalone Go module, server configuration, and tested HTTP skeleton | commit: pending.
- T2 | S1,S2,S4 | inline | pending | Define typed opportunity/plan contracts and deterministic eligibility, freshness, and evidence checks; cover the README scenarios with tests | commit: pending.
- T3 | S1,S2,S4 | inline | pending | Validate the current Google ADK Go API, then implement missing-information questions and evidence-grounded structured follow-up drafts with provider-independent tests | commit: pending.
- T4 | S1,S2,S3 | inline | pending | Build the browser workflow and local SQLite history for entering opportunities, reviewing/editing drafts, and revisiting prior plans | commit: pending.
- T5 | S1,S3,S4 | inline + delegated independent verifier | pending | Finish synthetic fixtures, privacy/security and error handling, user docs, full tests, and primary browser verification; resolve all findings | commit: pending.

## Log

- L1 — 2026-10-14 — Original request (verbatim):
```text
ok entonces vamos a construir desde cero este proyecto para que sea una app web con go y google adk y ademas crea el
 ▎ plan con odd y me avisas cuando tengas el plan listo
```
- L2 — 2026-10-14 — User direction (verbatim): `ok piensa como un agente puede ser mas util y usa la informacion que tienes`.
- L3 — 2026-10-14 — User constraint (verbatim): `el repo es publico`.
- L4 — 2026-10-14 — User confirmation (verbatim): `mira entonces vamos a crear este plan`.
- L5 — 2026-10-14 — Existing evidence: `README.md` documents a text-only, user-supplied opportunity workflow, typed `OpportunityInput`/`FollowUpPlan`, stop conditions for opt-outs/closed/stale opportunities, and no automatic email or CRM updates. The current `main.go` depends on the cookbook's `internal/adkhelper`, so the new application must be standalone.
- L6 — 2026-10-14 — Research evidence: the official ADK Go quickstart was found at https://google.github.io/adk-docs/get-started/go/, but fetching the page failed in this session. The exact module version and Go API remain unverified; T3 must validate them before implementation.
- L7 — 2026-10-14 — Route/risk: all implementation tasks are sequential and coupled, so they stay inline; no parallel-writer reason fired. Credentials, model-bound user data, and persistence make final independent verification necessary. The initial forecast is above 400 authored lines; retain `ask-on-risk` and resolve oversized-delivery strategy before a commit when the threshold policy requires it. No push, PR, merge, or public deployment is implied.
- L8 — 2026-10-14 — Current checkpoint: plan created on `feature/opportunity-follow-up-web-app` with five tasks; T1 is in progress. No source files have been changed yet. Continue within the original authorized build scope and stop for any unresolved product decision.
- L9 — 2026-10-14 — T1 RED: `go test ./internal/web` exited 1 with `internal/web/server_test.go:37:4: undefined: NewHandler`; the expected HTTP handler does not exist yet.
- L10 — 2026-10-14 — T1 RED: `go test ./internal/config` exited 1 with `internal/config/config_test.go:33:16: undefined: Load`; the expected configuration loader does not exist yet.
- L11 — 2026-10-14 — T1 GREEN: `go test ./internal/config ./internal/web` and `go test ./...` passed; `go vet ./...`, `git diff --check`, and `.gitignore` checks passed. The work-unit is ready but uncommitted; with a ~650-line feature forecast and `ask-on-risk`, its first commit waits for the user's oversized-delivery choice. No push or PR was created.
- L12 — 2026-10-14 — User selected `Cadena desde rama de feature (recomendada)`. Cache `chain_strategy=feature-branch-chain`; planned slices are PR1: T1-T2, PR2: T3, PR3: T4-T5, with exact commit boundaries to be recorded as work completes. This does not authorize push, PR creation, merge, or deployment.
