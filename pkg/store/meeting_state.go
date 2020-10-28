package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// The MeetingState holds a meeting and it's relations
// to a backend and frontend.
type MeetingState struct {
	ID string

	Meeting  *bbb.Meeting
	Frontend *FrontendState
	Backend  *BackendState

	CreatedAt time.Time
	UpdatedAt *time.Time
	SyncedAt  *time.Time

	pool *pgxpool.Pool
}

// InitMeetingState initializes meeting state with
// defaults and a connection
func InitMeetingState(
	pool *pgxpool.Pool,
	init *MeetingState,
) *MeetingState {
	init.pool = pool
	return init
}

// GetMeetingStates retrieves all meeting states
func GetMeetingStates(
	pool *pgxpool.Pool,
	q *Query,
) ([]*MeetingState, error) {
	ctx := context.Background()
	qry := `
		SELECT
		  id,

		  frontend_id,
		  backend_id,

		  state,

		  created_at,
		  updated_at,
		  synced_at
		FROM  meetings ` + q.related() + `
		WHERE ` + q.where()
	rows, err := pool.Query(ctx, qry, q.params()...)
	if err != nil {
		return nil, err
	}
	cmd := rows.CommandTag()
	results := make([]*MeetingState, 0, cmd.RowsAffected())
	for rows.Next() {
		state, err := meetingStateFromRow(pool, rows)
		if err != nil {
			return nil, err
		}
		results = append(results, state)
	}
	return results, nil
}

// GetMeetingState tries to retriev a single meeting state
func GetMeetingState(conn *pgxpool.Pool, q *Query) (*MeetingState, error) {
	states, err := GetMeetingStates(conn, q)
	if err != nil {
		return nil, err
	}
	if len(states) == 0 {
		return nil, nil
	}
	return states[0], nil
}

func meetingStateFromRow(
	conn *pgxpool.Pool,
	row pgx.Row,
) (*MeetingState, error) {
	state := InitMeetingState(conn, &MeetingState{})
	var (
		frontendID *string
		backendID  *string
	)

	err := row.Scan(
		&state.ID,
		&frontendID,
		&backendID,
		&state.Meeting,
		&state.CreatedAt,
		&state.UpdatedAt,
		&state.SyncedAt)
	if err != nil {
		return nil, err
	}

	// Get related backend state if present
	if backendID != nil {
		state.Backend, err = GetBackendState(conn, NewQuery().
			Eq("id", *backendID))
		if err != nil {
			return nil, err
		}
	}

	return state, err
}

// Refresh the backend state from the database
func (s *MeetingState) Refresh() error {
	// Load from database
	next, err := GetMeetingState(s.pool, NewQuery().
		Eq("id", s.ID))
	if err != nil {
		return err
	}
	*s = *next
	return nil
}

// Save updates or inserts a meeting state into our
// cluster state.
func (s *MeetingState) Save() error {
	var (
		err error
		id  string
	)
	if s.CreatedAt.IsZero() {
		id, err = s.insert()
		s.ID = id
	} else {
		err = s.update()
	}
	if err != nil {
		return err
	}

	return s.Refresh()
}

// Add a new meeting to the database
func (s *MeetingState) insert() (string, error) {
	ctx := context.Background()
	qry := `
		INSERT INTO meetings (
			id,
			state,

			backend_id,
			frontend_id,

			created_at,
			synced_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
}

func (s *MeetingState) update() error {
	return fmt.Errorf("implement me")
}
