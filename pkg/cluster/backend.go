package cluster

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
	"gitlab.com/infra.run/public/b3scale/pkg/store"
)

// A Backend is a BigBlueButton instance in the cluster.
//
// It has a bbb.backend secret for request authentication,
// stored in the backend state. The state is shared across all
// instances.
//
type Backend struct {
	state  *store.BackendState
	client *bbb.Client
	pool   *pgxpool.Pool
	cmds   *store.CommandQueue
}

// NewBackend creates a new backend instance with
// a fresh bbb client.
func NewBackend(pool *pgxpool.Pool, state *store.BackendState) *Backend {
	return &Backend{
		client: bbb.NewClient(),
		state:  state,
		pool:   pool,
		cmds:   store.NewCommandQueue(pool),
	}
}

// Backend State Sync: loadNodeState will make
// a small request to get a meeting that does not
// exist to check if the credentials are valid.
func (b *Backend) loadNodeState() error {
	log.Println(b.state.ID, "SYNC: node state")
	defer b.state.Save()

	// Measure latency
	t0 := time.Now()
	res, err := b.IsMeetingRunning(bbb.IsMeetingRunningRequest(
		bbb.Params{
			"meetingID": "00000000-0000-0000-0000-000000000001",
		}))
	t1 := time.Now()
	latency := t1.Sub(t0)

	if err != nil {
		errMsg := fmt.Sprintf("%s", err)
		b.state.NodeState = "error"
		b.state.LastError = &errMsg
		return errors.New(errMsg)
	}

	if res.Returncode != "SUCCESS" {
		// Update backend state
		errMsg := fmt.Sprintf("%s: %s", res.MessageKey, res.Message)
		b.state.LastError = &errMsg
		b.state.NodeState = "error"
		return errors.New(errMsg)
	}

	// Update state
	b.state.LastError = nil
	b.state.Latency = latency
	b.state.NodeState = "ready"
	return err
}

// Meeting State Sync: loadMeetingState will make
// a request to the backend with a get meeting info request
func (b *Backend) refreshMeetingState(
	state *store.MeetingState,
) error {
	req := bbb.GetMeetingInfoRequest(bbb.Params{
		"meetingID": state.ID,
	}).WithBackend(b.state.Backend)
	rep, err := b.client.Do(req)
	if err != nil {
		return err
	}
	res := rep.(*bbb.GetMeetingInfoResponse)
	if res.XMLResponse.Returncode != "SUCCESS" {
		return fmt.Errorf("meeting sync error: %v",
			res.XMLResponse.Message)
	}

	// Update meeting state
	state.Meeting = res.Meeting
	state.SyncedAt = time.Now().UTC()
	return state.Save()
}

// BBB API Implementation

func meetingStateFromRequest(
	pool *pgxpool.Pool,
	req *bbb.Request,
) (*store.MeetingState, error) {
	meetingID, ok := req.Params.MeetingID()
	if !ok {
		return nil, fmt.Errorf("meetingID required")
	}
	// Check if meeting does exist
	meetingState, err := store.GetMeetingState(pool, store.Q().
		Where("id = ?", meetingID))
	return meetingState, err
}

// Create a new Meeting
func (b *Backend) Create(req *bbb.Request) (
	*bbb.CreateResponse, error,
) {
	meetingState, err := meetingStateFromRequest(b.pool, req)
	if err != nil {
		return nil, err
	}
	if meetingState != nil {
		// Check if meeting is runnnig
		res, err := b.IsMeetingRunning(bbb.IsMeetingRunningRequest(
			bbb.Params{
				"meetingID": meetingState.ID,
			}))
		if err != nil {
			return nil, err
		}
		if res.XMLResponse.Returncode == "SUCCESS" {
			// We are good here, just return the current meeting
			// state in a synthetic response.
			res := &bbb.CreateResponse{
				XMLResponse: &bbb.XMLResponse{
					Returncode: "SUCCESS",
				},
				Meeting: meetingState.Meeting,
			}
			res.SetStatus(200)
			return res, nil
		}
	}

	// We don't know about the meeting, or is meeting
	// running did not know about the meeting - anyhow
	// we need to create it.
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	createRes := res.(*bbb.CreateResponse)
	// Update or save meeeting state
	if meetingState == nil {
		_, err = b.state.CreateMeetingState(req.Frontend, createRes.Meeting)
		if err != nil {
			return nil, err
		}
	} else {
		// Update state, associate with backend and frontend
		meetingState.Meeting = createRes.Meeting
		meetingState.SyncedAt = time.Now().UTC()
		if err := meetingState.Save(); err != nil {
			return nil, err
		}
	}

	return createRes, nil
}

// Join a meeting
func (b *Backend) Join(
	req *bbb.Request,
) (*bbb.JoinResponse, error) {
	// Joining a meeting is a process entirely handled by the
	// client. Because of a JSESSIONID which is used? I guess?
	// Just passing through the location response did not work.
	// For the reverseproxy feature we need to fix this.
	// Even if it means tracking JSESSIONID cookie headers.
	req = req.WithBackend(b.state.Backend)

	// Create custom join response
	res := &bbb.JoinResponse{
		XMLResponse: new(bbb.XMLResponse),
	}

	// TODO: Create HTML fallback for redirect
	res.SetStatus(http.StatusFound)
	res.SetRaw([]byte("<html><body>redirect fallback.</html>"))
	res.SetHeader(http.Header{
		"content-type": []string{"text/html"},
		"location":     []string{req.URL()},
	})

	// Dispatch updating the meeing state
	meetingID, _ := req.Params.MeetingID()
	b.cmds.Queue(UpdateMeetingState(&UpdateMeetingStateRequest{
		ID: meetingID,
	}))

	return res, nil
}

// IsMeetingRunning returns the is meeting running state
func (b *Backend) IsMeetingRunning(
	req *bbb.Request,
) (*bbb.IsMeetingRunningResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	isMeetingRunningRes := res.(*bbb.IsMeetingRunningResponse)
	meetingID, _ := req.Params.MeetingID()
	if isMeetingRunningRes.Returncode == "ERROR" {
		// Delete meeting
		store.DeleteMeetingState(b.pool, &store.MeetingState{ID: meetingID})
	}

	return isMeetingRunningRes, err
}

// End a meeting
func (b *Backend) End(req *bbb.Request) (*bbb.EndResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.EndResponse), err
}

// GetMeetingInfo gets the meeting details
func (b *Backend) GetMeetingInfo(
	req *bbb.Request,
) (*bbb.GetMeetingInfoResponse, error) {
	rep, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	res := rep.(*bbb.GetMeetingInfoResponse)

	// Update our meeting in the store
	if res.XMLResponse.Returncode == "SUCCESS" {
		meetingID, _ := req.Params.MeetingID()
		mstate, err := store.GetMeetingState(b.pool, store.Q().
			Where("id = ?", meetingID))
		if err != nil {
			// We only log the error, as this might fail
			// without impacting the service
			log.Println(err)
		} else {
			// Update meeting state
			mstate.Meeting = res.Meeting
			mstate.SyncedAt = time.Now().UTC()
			if err := mstate.Save(); err != nil {
				log.Println(err)
			}
		}
	}

	return res, nil
}

// GetMeetings retrieves a list of meetings
func (b *Backend) GetMeetings(
	req *bbb.Request,
) (*bbb.GetMeetingsResponse, error) {
	// Get all meetings from our store associated
	// with the requesting frontend.
	mstates, err := store.GetMeetingStates(b.pool, store.Q().
		Join("frontends ON frontends.id = meetings.frontend_id").
		Where("frontend.key = ?", req.Frontend))
	if err != nil {
		return nil, err
	}
	meetings := make([]*bbb.Meeting, 0, len(mstates))
	for _, state := range mstates {
		meetings = append(meetings, state.Meeting)
	}

	// Create response with all meetings
	res := &bbb.GetMeetingsResponse{
		XMLResponse: &bbb.XMLResponse{
			Returncode: "SUCCESS",
		},
		Meetings: meetings,
	}

	return res, nil
}

// GetRecordings retrieves a list of recordings
func (b *Backend) GetRecordings(
	req *bbb.Request,
) (*bbb.GetRecordingsResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.GetRecordingsResponse), nil
}

// PublishRecordings publishes a recording
func (b *Backend) PublishRecordings(
	req *bbb.Request,
) (*bbb.PublishRecordingsResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.PublishRecordingsResponse), nil
}

// DeleteRecordings deletes recordings
func (b *Backend) DeleteRecordings(
	req *bbb.Request,
) (*bbb.DeleteRecordingsResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.DeleteRecordingsResponse), nil
}

// UpdateRecordings updates recordings
func (b *Backend) UpdateRecordings(
	req *bbb.Request,
) (*bbb.UpdateRecordingsResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.UpdateRecordingsResponse), nil
}

// GetDefaultConfigXML retrieves the default config xml
func (b *Backend) GetDefaultConfigXML(
	req *bbb.Request,
) (*bbb.GetDefaultConfigXMLResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.GetDefaultConfigXMLResponse), nil
}

// SetConfigXML sets the? config xml
func (b *Backend) SetConfigXML(
	req *bbb.Request,
) (*bbb.SetConfigXMLResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.SetConfigXMLResponse), nil
}

// GetRecordingTextTracks retrievs all text tracks
func (b *Backend) GetRecordingTextTracks(
	req *bbb.Request,
) (*bbb.GetRecordingTextTracksResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.GetRecordingTextTracksResponse), nil
}

// PutRecordingTextTrack adds a text track
func (b *Backend) PutRecordingTextTrack(
	req *bbb.Request,
) (*bbb.PutRecordingTextTrackResponse, error) {
	res, err := b.client.Do(req.WithBackend(b.state.Backend))
	if err != nil {
		return nil, err
	}
	return res.(*bbb.PutRecordingTextTrackResponse), nil
}
