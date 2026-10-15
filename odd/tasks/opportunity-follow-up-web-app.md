# Opportunity Follow-up Web App — ODD Feature Plan
Objective: Rebuild the public prototype as a standalone Go web app using Google ADK for safe, evidence-grounded opportunity follow-up guidance.
Branch: `feature/opportunity-follow-up-web-app`; delivery strategy: `ask-on-risk` (resolved); chain strategy: `feature-branch-chain`; forecast: ~1,500 authored changed lines, to be re-estimated by work unit; no push, PR, or deployment authorized.
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

### S5 — Staleness rule
> Estado obsoleto explícito (recomendado)

> La persona marca los datos/etapa como obsoletos; la app no inventa un umbral temporal.

### Planning assumptions (not additional user requirements)
- Start as a local-only, server-rendered Go app; public source code does not imply public hosting.
- Use SQLite for local opportunity/draft history and synthetic demo fixtures; do not connect to a real CRM, calendar, or email sender in this scope.
- Keep Gemini credentials in server environment configuration; never commit or log them.
- Deterministic application rules enforce ineligibility before the model runs. The ADK agent asks for missing facts and drafts a typed plan; it does not decide opt-out/closed gates or perform external side effects.
- Confirm these defaults if implementation reveals a consequential product decision; no production data or deployment is assumed.

## Tasks

- T1 | S1,S3 | inline | done | Create a standalone Go module, server configuration, and tested HTTP skeleton | commit: e3d3efa46c39b5617a5398e5d6ac12acc9c07fd3.
- T2 | S1,S2,S4,S5 | inline | in_progress | Define typed opportunity/plan contracts and deterministic eligibility, freshness, and evidence checks; cover the README scenarios with tests; use separate T2a contracts/input-validation, T2b eligibility, and T2c plan-evidence-validation work-unit commits | commits: T2a 5a52afda79f6dbcd01d3895f67420d49eb4950de; T2b 983deb58b7ee570acb7fd91cfc9e9139e783ffac; T2c pending.
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
- L12 — 2026-10-14 — User selected `Cadena desde rama de feature (recomendada)`. Cache `chain_strategy=feature-branch-chain`; the initial planned slices were PR1: T1-T2, PR2: T3, PR3: T4-T5. This does not authorize push, PR creation, merge, or deployment.
- L13 — 2026-10-14 — T1 completed in commit `e3d3efa46c39b5617a5398e5d6ac12acc9c07fd3` (302 authored changed lines). Focused tests, `go test ./...`, `go vet ./...`, whitespace, and ignore-rule checks passed. Based on the observed size, revise provisional slices to PR1: T1; PR2: T2-T3; PR3: T4-T5; recheck authored-line totals before each slice.
- L14 — 2026-10-14 — T2 was blocked pending a product rule: the README requires stale data/stages to stop follow-up, but did not define how stale is determined.
- L15 — 2026-10-14 — User selected (verbatim): `Estado obsoleto explícito (recomendado)`; selected description (verbatim): `La persona marca los datos/etapa como obsoletos; la app no inventa un umbral temporal.` T2 is unblocked; require an explicit stale flag and do not infer staleness from age.
- L16 — 2026-10-14 — T2 RED: `go test ./internal/opportunity` exited 1 because `OpportunityInput`, `Assessment`, `StatusReady`, `Contact`, and other domain symbols do not yet exist; this is the expected pre-implementation failure.
- L17 — 2026-10-14 — T2 GREEN so far: `go test ./internal/opportunity`, `go test ./...`, and `go vet ./...` passed; `git diff --check` passed. The domain worktree is 631 Go lines across input/eligibility and generated-plan validation, so split it at natural behavior boundaries.
- L18 — 2026-10-14 — Refined T2 into three work units: T2a typed contracts/input validation, T2b eligibility assessment, T2c generated-plan/evidence validation. The feature forecast is revised to ~1,500 authored lines. Provisional chain slices: PR1 T1; PR2 T2a; PR3 T2b+T2c; PR4 T3; PR5 T4; PR6 T5; separate T2b/T2c or re-slice later if measured changes exceed the review budget. User-selected chain strategy remains `feature-branch-chain`; no push/PR/deployment is authorized.
- L19 — 2026-10-14 — After separating T2 tests and source by T2a/T2b/T2c, `go test ./internal/opportunity`, `go test ./...`, `go vet ./...`, and `git diff --check` passed. T2 remains in progress; no T2 work-unit commit has been made yet.
- L20 — 2026-10-14 — T2a isolated check passed: `go test internal/opportunity/types.go internal/opportunity/input_validation.go internal/opportunity/opportunity_test.go`. The T2a slice was 235 additions and 4 deletions (239 diff lines); T2b and T2c were kept outside it.
- L21 — 2026-10-14 — T2a committed as `5a52afda79f6dbcd01d3895f67420d49eb4950de` (`feat(domain): add opportunity input contracts`). T2 remains in progress while eligibility and plan/evidence validation are completed.
- L22 — 2026-10-14 — T2b isolated check passed: `go test internal/opportunity/types.go internal/opportunity/input_validation.go internal/opportunity/opportunity_test.go internal/opportunity/assessment.go internal/opportunity/assessment_test.go`. This validates the eligibility unit against committed T2a without including T2c plan validation.
- L23 — 2026-10-14 — T2b committed as `983deb58b7ee570acb7fd91cfc9e9139e783ffac` (`feat(domain): assess opportunity follow-up eligibility`). T2 remains in progress while plan/evidence validation is completed.
- L24 — 2026-10-14 — T2c full-domain isolated check passed: `go test internal/opportunity/types.go internal/opportunity/input_validation.go internal/opportunity/opportunity_test.go internal/opportunity/assessment.go internal/opportunity/assessment_test.go internal/opportunity/validation.go internal/opportunity/validation_test.go`; `go test ./...` and `go vet ./...` passed. T2 remains in progress until T2c's work-unit commit.
