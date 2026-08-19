package server

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
)

func (s *Server) Start(ctx context.Context) error {
	// TODO: first check here if already exists and running, if not then recreate, if yes resync state
	s.setStatus("starting")

	if err := s.Remove(ctx); err != nil {
		if !errdefs.IsNotFound(err) {
			s.setStatus("offline")
			s.bus.Publish(EventStateError, err.Error())
			return fmt.Errorf("failed container cleanup: %w", err)
		}
	}

	if err := s.Create(ctx); err != nil {
		s.setStatus("offline")
		s.bus.Publish(EventStateError, err.Error())
		return fmt.Errorf("failed container creation: %w", err)
	}

	sctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := s.Attach(sctx); err != nil {
		s.setStatus("offline")
		s.bus.Publish(EventStateError, err.Error())
		return fmt.Errorf("failed container attach: %w", err)
	}

	if _, err := s.cli.ContainerStart(sctx, s.id, client.ContainerStartOptions{}); err != nil {
		s.setStatus("offline")
		s.bus.Publish(EventStateError, err.Error())
		return fmt.Errorf("failed container start: %w", err)
	}

	// TODO: implementing a running detector or something
	s.setStatus("running")
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	// TODO: stop options
	s.setStatus("stopping")

	// TODO: make stop commands and stuff exist, too. this is just temporary
	if _, err := s.cli.ContainerStop(ctx, s.ID(), client.ContainerStopOptions{
		Timeout: new(60),
	}); err != nil && !errdefs.IsNotFound(err) {
		return err
	}

	if err := s.Remove(ctx); err != nil && !errdefs.IsNotFound(err) {
		return err
	}

	s.setStatus("offline")
	return nil
}

func (s *Server) Restart(ctx context.Context) error {
	if err := s.Stop(ctx); err != nil {
		return err
	}

	if err := s.Start(ctx); err != nil {
		return err
	}

	return nil
}
