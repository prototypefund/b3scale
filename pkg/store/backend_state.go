package store

import (
	"context"
	//	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// The BackendState is shared across b3scale instances
// and encapsulates the list of meetings and recordings.
// The backend.ID should be used as identifier.
type BackendState struct {
	ID string

	NodeState  string
	AdminState string

	LastError *string

	Backend *bbb.Backend

	Tags []string

	CreatedAt time.Time
	UpdatedAt *time.Time
	SyncedAt  *time.Time

	// DB
	conn *pgxpool.Pool
}

// InitBackendState initializes a new backend state with
// an initial state.
func InitBackendState(conn *pgxpool.Pool, init *BackendState) *BackendState {
	// Add default values
	if init.NodeState == "" {
		init.NodeState = "init"
	}
	if init.AdminState == "" {
		init.AdminState = "ready"
	}
	if init.Backend == nil {
		init.Backend = &bbb.Backend{}
	}
	if init.Tags == nil {
		init.Tags = []string{}
	}

	init.conn = conn
	return init
}

// GetBackendStates retrievs all backends
func GetBackendStates(conn *pgxpool.Pool, q *Query) ([]*BackendState, error) {
	ctx := context.Background()
	qry := `
		SELECT
		  id,

		  node_state,
		  admin_state,

		  last_error,

		  host,
		  secret,

		  tags,

		  created_at,
		  updated_at,
		  synced_at
		FROM backends ` + q.related() + `
		WHERE ` + q.where()
	rows, err := conn.Query(ctx, qry, q.params()...)
	if err != nil {
		return nil, err
	}
	cmd := rows.CommandTag()
	// fmt.Println("Affected rows:", cmd.RowsAffected())
	results := make([]*BackendState, 0, cmd.RowsAffected())
	for rows.Next() {
		state, err := backendStateFromRow(conn, rows)
		if err != nil {
			return nil, err
		}
		results = append(results, state)
	}

	return results, nil
}

// GetBackendState tries to retriev a single backend state
func GetBackendState(conn *pgxpool.Pool, q *Query) (*BackendState, error) {
	states, err := GetBackendStates(conn, q)
	if err != nil {
		return nil, err
	}
	if len(states) == 0 {
		return nil, nil
	}
	return states[0], nil
}

func backendStateFromRow(conn *pgxpool.Pool, row pgx.Row) (*BackendState, error) {
	state := InitBackendState(conn, &BackendState{})
	err := row.Scan(
		&state.ID,
		&state.NodeState,
		&state.AdminState,
		&state.LastError,
		&state.Backend.Host,
		&state.Backend.Secret,
		&state.Tags,
		&state.CreatedAt,
		&state.UpdatedAt,
		&state.SyncedAt)
	return state, err
}

// Refresh the backend state from the database
func (s *BackendState) Refresh() error {
	// Load from database
	q := NewQuery().Eq("id", s.ID)
	next, err := GetBackendState(s.conn, q)
	if err != nil {
		return err
	}
	*s = *next
	return nil
}

// Save persists the backend state in the database store
func (s *BackendState) Save() error {
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

// Private insert: adds a new row to the backends table
func (s *BackendState) insert() (string, error) {
	ctx := context.Background()
	qry := `
		INSERT INTO backends (
			host,
			secret,

			node_state,
			admin_state,

			tags
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	insertID := ""
	err := s.conn.QueryRow(ctx, qry,
		// Values
		s.Backend.Host,
		s.Backend.Secret,
		s.NodeState,
		s.AdminState,
		s.Tags).Scan(&insertID)

	return insertID, err
}

// Private update: updates the db row
func (s *BackendState) update() error {
	now := time.Now().UTC()
	s.UpdatedAt = &now
	ctx := context.Background()
	qry := `
		UPDATE backends
		   SET node_state   = $2,
		       admin_state  = $3,

			   last_error   = $4,

			   host         = $5,
			   secret       = $6,

			   tags         = $7,

			   updated_at   = $8,
			   synced_at    = $9

		 WHERE id = $1
	`
	_, err := s.conn.Exec(
		ctx, qry,
		// Identifier
		s.ID,
		// Update Values
		s.NodeState,
		s.AdminState,
		s.LastError,
		s.Backend.Host,
		s.Backend.Secret,
		s.Tags,
		s.UpdatedAt,
		s.SyncedAt)

	return err
}

// ClearMeetings will remove all meetings in the current state
func (s *BackendState) ClearMeetings() error {
	ctx := context.Background()
	qry := `
		DELETE FROM meetings WHERE backend_id = $1
	`
	_, err := s.conn.Exec(ctx, qry, s.ID)
	return err
}

// SetMeetings will replace all meetings in the current state
func (s *BackendState) SetMeetings(meetings []*bbb.Meeting) error {
	ctx := context.Background()
	tx, err := s.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// First clear all meetings
	if err := s.ClearMeetings(); err != nil {
		return err
	}

	// The add all new states
	for _, m := range meetings {
		// Create meeting state
		state := InitMeetingState(s.conn, &MeetingState{
			Backend: s,
			Meeting: m,
		})
		state.Save()
	}

	now := time.Now().UTC()
	s.NodeState = "ready"
	s.SyncedAt = &now
	s.UpdatedAt = &now
	s.Save()

	// We are done here
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

// GetMeetingStates retrievs all meeting states for
// a given backend state.
func (s *BackendState) GetMeetingStates(q *Query) ([]*MeetingState, error) {
	return nil, nil
}
