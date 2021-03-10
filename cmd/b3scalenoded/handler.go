package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog/log"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// The EventHandler processes BBB Events and updates
// the cluster state
type EventHandler struct {
	backend *store.BackendState
}

// NewEventHandler creates a new handler instance
// with a database pool
func NewEventHandler(backend *store.BackendState) *EventHandler {
	return &EventHandler{
		backend: backend,
	}
}

// Dispatch invokes the handler functions on the BBB event
func (h *EventHandler) Dispatch(ctx context.Context, e bbb.Event) error {
	tx, err := store.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	switch e.(type) {
	case *bbb.MeetingCreatedEvent:
		err = h.onMeetingCreated(ctx, tx, e.(*bbb.MeetingCreatedEvent))
	case *bbb.MeetingEndedEvent:
		err = h.onMeetingEnded(ctx, tx, e.(*bbb.MeetingEndedEvent))
	case *bbb.MeetingDestroyedEvent:
		err = h.onMeetingDestroyed(ctx, tx, e.(*bbb.MeetingDestroyedEvent))

	case *bbb.UserJoinedMeetingEvent:
		err = h.onUserJoinedMeeting(ctx, tx, e.(*bbb.UserJoinedMeetingEvent))
	case *bbb.UserLeftMeetingEvent:
		err = h.onUserLeftMeeting(ctx, tx, e.(*bbb.UserLeftMeetingEvent))

	default:
		log.Error().
			Str("type", fmt.Sprintf("%T", e)).
			Str("event", fmt.Sprintf("%v", e)).
			Msg("unhandled unknown event")
	}
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// handle event: MeetingCreated
func (h *EventHandler) onMeetingCreated(
	ctx context.Context,
	tx pgx.Tx,
	e *bbb.MeetingCreatedEvent,
) error {
	log.Info().
		Str("internalMeetingID", e.InternalMeetingID).
		Str("meetingID", e.MeetingID).
		Msg("meeting created")

	// TODO this might be handled through the context...
	deadline := 5 * time.Second
	mstate, err := awaitInternalMeeting(
		ctx, tx, e.InternalMeetingID, deadline)
	if err != nil {
		return err
	}

	if mstate == nil {
		log.Warn().
			Str("internalMeetingID", e.InternalMeetingID).
			Str("meetingID", e.MeetingID).
			Msg("meeting identified by internalMeetingID " +
				"is unknown  to the cluster")
		return nil
	}
	mstate.Meeting.Running = true
	if err := mstate.Save(ctx, tx); err != nil {
		return err
	}

	return nil
}

// handle event: MeetingEnded
func (h *EventHandler) onMeetingEnded(
	ctx context.Context,
	tx pgx.Tx,
	e *bbb.MeetingEndedEvent,
) error {
	log.Info().
		Str("internalMeetingID", e.InternalMeetingID).
		Msg("meeting ended")
	deadline := 5 * time.Second
	mstate, err := awaitInternalMeeting(
		ctx, tx, e.InternalMeetingID, deadline)
	if err != nil {
		return err
	}
	if mstate == nil {
		log.Warn().
			Str("internalMeetingID", e.InternalMeetingID).
			Msg("meeting identified by internalMeetingID " +
				"is unknown to the cluster")
	}
	// Reset meeting state
	mstate.Meeting.Running = false
	mstate.Meeting.Attendees = []*bbb.Attendee{}
	if err := mstate.Save(ctx, tx); err != nil {
		return err
	}
	return nil
}

// handle event: MeetingDestroyed
func (h *EventHandler) onMeetingDestroyed(
	ctx context.Context,
	tx pgx.Tx,
	e *bbb.MeetingDestroyedEvent,
) error {
	log.Info().
		Str("internalMeetingID", e.InternalMeetingID).
		Msg("meeting destroyed")
	// Delete meeting state
	err := store.DeleteMeetingStateByInternalID(ctx, tx, e.InternalMeetingID)
	if err != nil {
		return nil
	}

	return nil
}

// handle event: UserJoinedMeeting
func (h *EventHandler) onUserJoinedMeeting(
	ctx context.Context,
	tx pgx.Tx,
	e *bbb.UserJoinedMeetingEvent,
) error {
	log.Info().
		Str("userID", e.Attendee.InternalUserID).
		Str("userFullName", e.Attendee.FullName).
		Str("internalMeetingID", e.InternalMeetingID).
		Msg("user joined meeting")

	// Insert (preliminar) attendee into the meeting state
	mstate, err := store.GetMeetingState(ctx, tx, store.Q().
		Where("meetings.internal_id = ?", e.InternalMeetingID))
	if err != nil {
		return err
	}

	if mstate == nil {
		log.Warn().
			Str("internalMeetingID", e.InternalMeetingID).
			Msg("meeting identified by internalMeetingID " +
				"is unknown to the cluster")
		return nil // however we are done here
	}

	// Update state attendees list
	if mstate.Meeting.Attendees == nil {
		mstate.Meeting.Attendees = []*bbb.Attendee{}
	}
	mstate.Meeting.Attendees = append(
		mstate.Meeting.Attendees,
		e.Attendee)

	if err := mstate.Save(ctx, tx); err != nil {
		return err
	}

	return nil
}

// handle event: UserLeftMeeting
func (h *EventHandler) onUserLeftMeeting(
	ctx context.Context,
	tx pgx.Tx,
	e *bbb.UserLeftMeetingEvent,
) error {
	log.Info().
		Str("internalUserID", e.InternalUserID).
		Str("internalMeetingID", e.InternalMeetingID).
		Msg("user left meeting")

	// Remove user from attendees list
	mstate, err := store.GetMeetingState(ctx, tx, store.Q().
		Where("meetings.internal_id = ?", e.InternalMeetingID))
	if err != nil {
		return err
	}

	if mstate == nil {
		log.Warn().
			Str("internalMeetingID", e.InternalMeetingID).
			Msg("meeting identified by internalMeetingID " +
				"is unknown to the cluster")
		return nil // however we are done here
	}

	// Update state attendees list
	if mstate.Meeting.Attendees == nil {
		return nil // Unlikely but in this case, we are done here
	}

	// Remove user from meeting's attendees
	filtered := make([]*bbb.Attendee, 0, len(mstate.Meeting.Attendees))
	for _, a := range mstate.Meeting.Attendees {
		if a.InternalUserID == e.InternalUserID {
			continue // The user just left
		}
		filtered = append(filtered, a)
	}
	mstate.Meeting.Attendees = filtered
	if err := mstate.Save(ctx, tx); err != nil {
		return err
	}

	return nil
}
