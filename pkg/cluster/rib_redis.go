package cluster

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"

	"gitlab.com/infra.run/public/b3scale/pkg/bbb"
)

// Create a meeting key for a frontend value
func frontendMeetingKey(id string) string {
	return "m:" + id + ":fe"
}

// Create a meeting key for a backend value
func backendMeetingKey(id string) string {
	return "m:" + id + ":be"
}

// A RedisRIB is an implementation of the RIB interface
// using the redis key value store as stand alone or
// behind a HAProxy or something.
type RedisRIB struct {
	state *State
	rdb   *redis.Client
}

// NewRedisRIB creates a new redis client and
// uses the state provided for retrieving backends
// and frontends by ID.
func NewRedisRIB(state *State, opts *redis.Options) *RedisRIB {
	rdb := redis.NewClient(opts)
	return &RedisRIB{
		state: state,
		rdb:   rdb,
	}
}

// SetBackend associates a backend id with a meeting.
// When b is nil the key will be deleted.
func (rib *RedisRIB) SetBackend(m *bbb.Meeting, b *Backend) error {
	// Check identifiers
	if m.MeetingID == "" {
		return fmt.Errorf("meeting id is empty")
	}

	ctx := context.Background()
	key := backendMeetingKey(m.MeetingID)

	// Is this a delete operation?
	if b == nil {
		if err := rib.rdb.Del(ctx, key).Err(); !errors.Is(err, redis.Nil) {
			return err
		}
		return nil // We are done here
	}

	// Otherwise set backend, if possible
	if b.ID == "" {
		return fmt.Errorf("backend id is empty")
	}
	return rib.rdb.Set(ctx, key, b.ID, 0).Err()
}

// GetBackend retrieves a backend from the store
// or returns nil in case it could not be found.
func (rib *RedisRIB) GetBackend(m *bbb.Meeting) (*Backend, error) {
	if m.MeetingID == "" {
		return nil, fmt.Errorf("meeting id is empty")
	}

	ctx := context.Background()
	key := backendMeetingKey(m.MeetingID)

	// Lookup backend
	id, err := rib.rdb.Get(ctx, key).Result()
	if err != nil {
		// Ignore if the key was not found
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	backend := rib.state.GetBackendByID(id)
	return backend, nil
}

// Delete is removing the meeting from the store
func (rib *RedisRIB) Delete(m *bbb.Meeting) error {
	err := rib.SetBackend(m, nil)
	if err != nil {
		return err
	}

	return err
}

// A RedisClusterRIB is a implemenation of a RIB
// interface using a redis cluster client.
type RedisClusterRIB struct {
	state *State
	rdb   *redis.ClusterClient
}

// NewRedisClusterRIB makes a new store with
// a redis host address and a cluster state.
func NewRedisClusterRIB(
	state *State,
	opts *redis.ClusterOptions,
) *RedisClusterRIB {
	rdb := redis.NewClusterClient(opts)
	return &RedisClusterRIB{
		state: state,
		rdb:   rdb,
	}
}

// GetBackend retrieves a backend associated with
// a Meeting from the store.
// If no meeting was found, nil will be returned.
func (s *RedisClusterRIB) GetBackend(
	meeting *bbb.Meeting,
) (*Backend, error) {
	// Check identifier
	if meeting.MeetingID == "" {
		return nil, fmt.Errorf("meeting id is empty")
	}
	ctx := context.Background()
	// Lookup backend id in cache
	id, err := s.rdb.Get(ctx, meeting.MeetingID).Result()
	if err != nil {
		return nil, err
	}
	if id == "" {
		return nil, nil
	}
	backend := s.state.GetBackendByID(id)
	return backend, nil
}

// SetBackend associates a backend with a meeting
func (s *RedisClusterRIB) SetBackend(
	meeting *bbb.Meeting, backend *Backend,
) error {
	// Check identifiers
	if meeting.MeetingID == "" {
		return fmt.Errorf("meeting id is empty")
	}
	if backend.ID == "" {
		return fmt.Errorf("backend id is empty")
	}

	ctx := context.Background()
	return s.rdb.Set(ctx, meeting.MeetingID, backend.ID, 0).Err()
}