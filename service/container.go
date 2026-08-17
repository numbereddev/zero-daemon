package service

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func (s *Service) checkImage(image string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// TODO: for now it just always pulls but we gotta make this smarter

	reader, err := s.client.ImagePull(ctx, image, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed image pull: %w", err)
	}

	defer func() { _ = reader.Close() }()
	// TODO: make it actually parse the lines and send it properly in
	_, _ = io.Copy(s.eventBus, reader)

	return nil
}

func (s *Service) Start(ctx context.Context) error {
	s.SetState("running")

	if err := s.Attach(ctx); err != nil {
		return fmt.Errorf("failed container attach: %w", err)
	}

	if _, err := s.client.ContainerStart(ctx, s.id, client.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed container start: %w", err)
	}

	return nil
}

func (s *Service) Create(ctx context.Context) error {
	image := "docker.io/library/alpine"
	if err := s.checkImage(image); err != nil {
		return fmt.Errorf("failed pulling image: %w", err)
	}

	_, err := s.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		// This is okay, because Docker lets us do everything on containers with both the Docker ID
		// or the container name interchangeably.
		Name: s.ID(),
		Config: &container.Config{
			Cmd: []string{"echo", "hello, world\n"},
			Tty: false,
		},
		Image: image,
	})
	if err != nil {
		return fmt.Errorf("failed container create: %w", err)
	}

	return nil
}

func (s *Service) Attach(ctx context.Context) error {
	if s.IsAttached() {
		return nil
	}

	if stream, err := s.client.ContainerAttach(ctx, s.ID(), client.ContainerAttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	}); err != nil {
		return fmt.Errorf("failed container attach: %w", err)
	} else {
		s.setStream(&stream.HijackedResponse)
	}

	go func() {
		defer s.stream.Close()
		defer func() { s.setStream(nil) }()

		// TODO: do something else, event system shouldn't be used for this stuff.
		_, _ = io.Copy(s.eventBus, s.stream.Reader)
	}()

	return nil
}

// TODO: Add Remove options
func (s *Service) Remove(ctx context.Context) error {
	if _, err := s.client.ContainerRemove(ctx, s.ID(), client.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		return err
	}

	return nil
}
