package service

import (
	"context"
	"sync"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
	"github.com/numbereddev/zero-daemon/events"
)

type Service struct {
	mu     sync.RWMutex
	client *client.Client

	id string

	// TODO: needs to be reworked, just temp
	state   string
	stateMx sync.Mutex

	stream   *client.HijackedResponse
	eventBus *events.Bus[[]byte]
}

// TODO: Configuration
func New(cli *client.Client, id string) *Service {
	return &Service{
		id:       id,
		client:   cli,
		eventBus: events.NewBus[[]byte](),
	}
}

func (s *Service) ID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.id
}

func (s *Service) Events() events.Reader[[]byte] {
	return s.eventBus
}

func (s *Service) IsAttached() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.stream != nil
}

func (s *Service) setStream(stream *client.HijackedResponse) {
	s.mu.Lock()
	s.stream = stream
	s.mu.Unlock()
}

func (s *Service) State() string {
	s.stateMx.Lock()
	defer s.stateMx.Unlock()
	return s.state
}

func (s *Service) SetState(action string) {
	s.stateMx.Lock()
	s.state = action
	s.stateMx.Unlock()
}

func (s *Service) Exists() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := s.client.ContainerInspect(ctx, s.ID(), client.ContainerInspectOptions{}); err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
