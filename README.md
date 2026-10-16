# Opportunity Follow-up

A standalone, single-user Go web app for preparing evidence-grounded opportunity follow-ups from information a sales owner supplies. It checks eligibility, asks for missing details, and—when configured—uses the Google Agent Development Kit (ADK) for Go with Google Gemini to create a structured draft for the owner to review and edit.

## Safety boundaries

- The web server accepts only loopback addresses. The default is `127.0.0.1:8080`; this app is not intended to be hosted or exposed to a network.
- Opted-out contacts, closed opportunities, and opportunities explicitly marked stale are blocked. The app does not infer staleness from age.
- Drafts are suggestions for human review. The app does not send email, read or update a CRM, change pipeline state, access a calendar, or support multiple users.
- If a Google API key is configured and an opportunity is eligible, its supplied ID, stage, facts, contact details, last-contact text, and goal are sent to Google Gemini to generate a draft. The key stays on the server and is not included in the page.

## Quick start

Requirements: Go 1.26.6 or newer.

```bash
export GOOGLE_API_KEY="your-key" # Optional; required to generate drafts
# Optional: export GEMINI_MODEL="your-model"
go run .
```

Open <http://127.0.0.1:8080>. The server binds to loopback by default. Without `GOOGLE_API_KEY`, it still starts and can assess and save submissions, but it cannot generate a draft.

Keep API keys in the server's environment. Do not commit them or enter them into the browser.

## Configuration

Set these environment variables before starting the app:

| Variable | Default | Purpose |
| --- | --- | --- |
| `APP_ADDR` | `127.0.0.1:8080` | HTTP listen address. The host must be loopback (`localhost` or a loopback IP) and the port must be valid. |
| `HISTORY_DB_PATH` | `data/opportunity-follow-up.sqlite` | Local SQLite history file. Relative paths are resolved from the working directory. |
| `GOOGLE_API_KEY` | empty | Server-side Google API key. Draft generation is unavailable when unset. |
| `GEMINI_MODEL` | `gemini-flash-latest` | Gemini model used when generation is enabled. |

For example, to use another local port, set `APP_ADDR=127.0.0.1:8081`. Non-loopback addresses are rejected.

## Use the app

1. Enter an opportunity ID, stage, contact, last-contact text, and goal.
2. Add evidence facts one per line in the format `fact-id | fact text`.
3. Submit the form. The app reports missing information, blocks opted-out/closed/owner-marked-stale opportunities, or prepares a draft when the opportunity is eligible and generation is configured.
4. Review the evidence references and edit the draft message or next action as needed. Nothing is sent.
5. Revisit saved submissions from **History**.

Use fictional data for evaluation. The app saves valid submissions locally, including records that need more information or are blocked.

## Local history and privacy

Opportunity inputs, assessments, and editable drafts are stored in SQLite on the machine running the app. The default database is `data/opportunity-follow-up.sqlite`; database files and the default `data/` directory are excluded from Git. For a new database path, the app creates parent directories with mode `0700` and sets the database file to mode `0600` on systems that honor POSIX permissions. The database is not encrypted by the app.

The current version has no in-app delete control or automatic expiration. Records remain in the database until it is removed manually. Stop the app before removing the file at the configured `HISTORY_DB_PATH` (or the default path above). The retention/deletion policy is not yet finalized, so protect the host account and database accordingly.

When generation is enabled, only eligible submissions are sent to Gemini. The request includes the opportunity ID and stage, supplied facts, contact ID and name, last-contact text, and goal. Review your organization's data-handling rules before using real customer or contact information.

## Development checks

```bash
go test ./...
go build .
```

The automated tests use deterministic generators and do not make live Gemini calls. No deployment or external CRM, email, calendar, or pipeline integration is included.
