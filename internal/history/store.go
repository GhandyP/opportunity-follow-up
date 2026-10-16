package history

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/GhandyP/opportunity-follow-up/internal/agent"
	"github.com/GhandyP/opportunity-follow-up/internal/opportunity"
	_ "modernc.org/sqlite"
)

var (
	ErrStoreUnavailable     = errors.New("history store is unavailable")
	ErrDatabasePathRequired = errors.New("history database path is required")
	ErrInvalidDraft         = errors.New("draft message and next action are required")
	ErrNotFound             = errors.New("history record not found")
	ErrDraftUnavailable     = errors.New("history record has no editable draft")
	ErrCorruptRecord        = errors.New("history record is corrupt")
)

type Record struct {
	ID              int64
	Input           opportunity.OpportunityInput
	Assessment      opportunity.Assessment
	Plan            *opportunity.FollowUpPlan
	DraftMessage    string
	DraftNextAction string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, path string) (*Store, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, ErrDatabasePathRequired
	}

	if path != ":memory:" {
		if dir := filepath.Dir(path); dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o700); err != nil {
				return nil, fmt.Errorf("create history directory: %w", err)
			}
		}
		file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
		if err != nil {
			return nil, fmt.Errorf("open history database file: %w", err)
		}
		if err := file.Chmod(0o600); err != nil {
			_ = file.Close()
			return nil, fmt.Errorf("protect history database file: %w", err)
		}
		if err := file.Close(); err != nil {
			return nil, fmt.Errorf("close history database file: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open SQLite history: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	store := &Store{db: db}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect to SQLite history: %w", err)
	}
	if err := store.initialize(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) initialize(ctx context.Context) error {
	if s == nil || s.db == nil {
		return ErrStoreUnavailable
	}
	const schema = `
CREATE TABLE IF NOT EXISTS follow_up_records (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	input_json TEXT NOT NULL,
	assessment_json TEXT NOT NULL,
	plan_json TEXT,
	draft_message TEXT NOT NULL DEFAULT '',
	draft_next_action TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
)`
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("initialize SQLite history: %w", err)
	}
	return nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Create(ctx context.Context, input opportunity.OpportunityInput, result agent.Result) (Record, error) {
	if s == nil || s.db == nil {
		return Record{}, ErrStoreUnavailable
	}
	expected, err := opportunity.AssessOpportunity(input)
	if err != nil {
		return Record{}, err
	}
	if !reflect.DeepEqual(expected, result.Assessment) {
		return Record{}, opportunity.ErrAssessmentMismatch
	}
	if result.Plan != nil {
		if err := opportunity.ValidateFollowUpPlan(input, result.Assessment, *result.Plan); err != nil {
			return Record{}, err
		}
	}

	inputJSON, err := json.Marshal(input)
	if err != nil {
		return Record{}, fmt.Errorf("encode opportunity history: %w", err)
	}
	assessmentJSON, err := json.Marshal(result.Assessment)
	if err != nil {
		return Record{}, fmt.Errorf("encode opportunity assessment: %w", err)
	}

	var planJSON any
	draftMessage, draftNextAction := "", ""
	var savedPlan *opportunity.FollowUpPlan
	if result.Plan != nil {
		encodedPlan, err := json.Marshal(result.Plan)
		if err != nil {
			return Record{}, fmt.Errorf("encode opportunity draft: %w", err)
		}
		planJSON = string(encodedPlan)
		draftMessage = result.Plan.Message
		draftNextAction = result.Plan.NextAction
		planCopy := *result.Plan
		savedPlan = &planCopy
	}

	now := time.Now().UTC()
	timestamp := now.Format(time.RFC3339Nano)
	insert, err := s.db.ExecContext(ctx, `
INSERT INTO follow_up_records (
	input_json, assessment_json, plan_json, draft_message, draft_next_action, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		string(inputJSON), string(assessmentJSON), planJSON, draftMessage, draftNextAction, timestamp, timestamp)
	if err != nil {
		return Record{}, fmt.Errorf("save opportunity history: %w", err)
	}
	id, err := insert.LastInsertId()
	if err != nil {
		return Record{}, fmt.Errorf("identify saved opportunity history: %w", err)
	}

	return Record{
		ID:              id,
		Input:           input,
		Assessment:      result.Assessment,
		Plan:            savedPlan,
		DraftMessage:    draftMessage,
		DraftNextAction: draftNextAction,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (s *Store) List(ctx context.Context) ([]Record, error) {
	if s == nil || s.db == nil {
		return nil, ErrStoreUnavailable
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, input_json, assessment_json, plan_json, draft_message, draft_next_action, created_at, updated_at
FROM follow_up_records ORDER BY id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list opportunity history: %w", err)
	}
	defer rows.Close()

	records := make([]Record, 0)
	for rows.Next() {
		record, err := scanRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read opportunity history: %w", err)
	}
	return records, nil
}

func (s *Store) Get(ctx context.Context, id int64) (Record, error) {
	if s == nil || s.db == nil {
		return Record{}, ErrStoreUnavailable
	}
	if id < 1 {
		return Record{}, ErrNotFound
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id, input_json, assessment_json, plan_json, draft_message, draft_next_action, created_at, updated_at
FROM follow_up_records WHERE id = ?`, id)
	return scanRecord(row)
}

func (s *Store) UpdateDraft(ctx context.Context, id int64, message, nextAction string) (Record, error) {
	if s == nil || s.db == nil {
		return Record{}, ErrStoreUnavailable
	}
	if id < 1 {
		return Record{}, ErrNotFound
	}
	if strings.TrimSpace(message) == "" || strings.TrimSpace(nextAction) == "" {
		return Record{}, ErrInvalidDraft
	}

	result, err := s.db.ExecContext(ctx, `
UPDATE follow_up_records
SET draft_message = ?, draft_next_action = ?, updated_at = ?
WHERE id = ? AND plan_json IS NOT NULL`,
		message, nextAction, time.Now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return Record{}, fmt.Errorf("update opportunity draft: %w", err)
	}
	rowsChanged, err := result.RowsAffected()
	if err != nil {
		return Record{}, fmt.Errorf("check opportunity draft update: %w", err)
	}
	if rowsChanged == 0 {
		var planExists int
		err := s.db.QueryRowContext(ctx, `SELECT CASE WHEN plan_json IS NULL THEN 0 ELSE 1 END FROM follow_up_records WHERE id = ?`, id).Scan(&planExists)
		if errors.Is(err, sql.ErrNoRows) {
			return Record{}, ErrNotFound
		}
		if err != nil {
			return Record{}, fmt.Errorf("check opportunity draft: %w", err)
		}
		if planExists == 0 {
			return Record{}, ErrDraftUnavailable
		}
	}
	return s.Get(ctx, id)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanRecord(row scanner) (Record, error) {
	var record Record
	var inputJSON, assessmentJSON, createdAt, updatedAt string
	var planJSON sql.NullString
	if err := row.Scan(
		&record.ID,
		&inputJSON,
		&assessmentJSON,
		&planJSON,
		&record.DraftMessage,
		&record.DraftNextAction,
		&createdAt,
		&updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Record{}, ErrNotFound
		}
		return Record{}, fmt.Errorf("read opportunity history record: %w", err)
	}
	if err := json.Unmarshal([]byte(inputJSON), &record.Input); err != nil {
		return Record{}, fmt.Errorf("%w: decode opportunity input: %v", ErrCorruptRecord, err)
	}
	if err := json.Unmarshal([]byte(assessmentJSON), &record.Assessment); err != nil {
		return Record{}, fmt.Errorf("%w: decode assessment: %v", ErrCorruptRecord, err)
	}
	if planJSON.Valid {
		var plan opportunity.FollowUpPlan
		if err := json.Unmarshal([]byte(planJSON.String), &plan); err != nil {
			return Record{}, fmt.Errorf("%w: decode draft: %v", ErrCorruptRecord, err)
		}
		record.Plan = &plan
	}
	var err error
	if record.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt); err != nil {
		return Record{}, fmt.Errorf("%w: parse creation time: %v", ErrCorruptRecord, err)
	}
	if record.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt); err != nil {
		return Record{}, fmt.Errorf("%w: parse update time: %v", ErrCorruptRecord, err)
	}
	return record, nil
}
