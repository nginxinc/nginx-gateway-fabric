package agent

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"
)

// ConnectInfoStore stores Agent ConnectInfo.
//
// Agents only send the ConnectInfo once on startup, and in the event of a reconnect,
// we need to be able to recover the ConnectInfo. However, we don't want to persist ConnectInfo
// for Agents that are no longer running.
//
// The ConnectInfoStore handles this by accepting a ttl, which is the minimum amount of time that
// ConnectInfo entries live after deletion.
//
// When ConnectInfo is deleted from the store, it will be marked for deletion, but will still be accessible until the
// ttl has been reached. During this time, if the ConnectInfo is requested, it will be unmarked for deletion.
type ConnectInfoStore struct {
	deleted            map[string]*entry
	entries            map[string]*entry
	logger             logr.Logger
	collectionInterval time.Duration
	ttl                time.Duration

	// mu protects both maps
	mu sync.Mutex
}

type entry struct {
	deletionTimestamp time.Time
	info              ConnectInfo
}

// NewConnectInfoStore returns a ConnectInfoStore with the provided ttl.
func NewConnectInfoStore(logger logr.Logger, ttl time.Duration) *ConnectInfoStore {
	return &ConnectInfoStore{
		deleted:            make(map[string]*entry),
		entries:            make(map[string]*entry),
		collectionInterval: ttl,
		ttl:                ttl,
		logger:             logger,
	}
}

// Start kicks off the garbage collection job for the ConnectInfoStore.
func (s *ConnectInfoStore) Start(ctx context.Context) error {
	ticker := time.NewTicker(s.collectionInterval)

	s.logger.Info("Staring garbage collection job for agent store")

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping garbage collection job for agent store")
			return nil
		case <-ticker.C:
			s.logger.Info("Checking for expired entries")
			s.mu.Lock()
			for id, entry := range s.deleted {
				if time.Since(entry.deletionTimestamp) > s.ttl {
					s.logger.Info("Deleting entry", "id", id)
					delete(s.deleted, id)
				}
			}
			s.mu.Unlock()
		}
	}
}

func (s *ConnectInfoStore) Get(id string) (ConnectInfo, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.entries[id]
	if ok {
		return entry.info, ok
	}

	entry, ok = s.deleted[id]
	if ok {
		delete(s.deleted, id)
		s.entries[id] = entry
		return entry.info, ok
	}

	return ConnectInfo{}, false
}

func (s *ConnectInfoStore) Add(info ConnectInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[info.ID] = &entry{info: info}
}

func (s *ConnectInfoStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, ok := s.entries[id]
	if !ok {
		return
	}

	entry.deletionTimestamp = time.Now()

	s.deleted[id] = entry
	delete(s.entries, id)
}
