package runtime

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func (r *Runtime) checkImage(image string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// TODO: for now it just always pulls but we gotta make this smarter

	reader, err := r.client.ImagePull(ctx, image, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed image pull: %w", err)
	}

	defer func() { _ = reader.Close() }()
	// TODO: make it actually parse the lines and send it properly in
	_, _ = io.Copy(r.eventBus, reader)

	return nil
}

func (r *Runtime) Start(ctx context.Context) error {
	r.SetState("running")

	if err := r.Attach(ctx); err != nil {
		return fmt.Errorf("failed container attach: %w", err)
	}

	if _, err := r.client.ContainerStart(ctx, r.id, client.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed container start: %w", err)
	}

	return nil
}

func (r *Runtime) Create(ctx context.Context) error {
	image := "docker.io/library/alpine"
	if err := r.checkImage(image); err != nil {
		return fmt.Errorf("failed pulling image: %w", err)
	}

	_, err := r.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		// This is okay, because Docker lets us do everything on containers with both the Docker ID
		// or the container name interchangeably.
		Name: r.ID(),
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

func (r *Runtime) Attach(ctx context.Context) error {
	if r.IsAttached() {
		return nil
	}

	if stream, err := r.client.ContainerAttach(ctx, r.ID(), client.ContainerAttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	}); err != nil {
		return fmt.Errorf("failed container attach: %w", err)
	} else {
		r.setStream(&stream.HijackedResponse)
	}

	go func() {
		defer r.stream.Close()
		defer func() { r.setStream(nil) }()

		// TODO: do something else, event system shouldn't be used for this stuff.
		_, _ = io.Copy(r.eventBus, r.stream.Reader)
	}()

	return nil
}

// TODO: Add Remove options
func (r *Runtime) Remove(ctx context.Context) error {
	if _, err := r.client.ContainerRemove(ctx, r.ID(), client.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		return err
	}

	return nil
}
