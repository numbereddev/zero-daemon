package container

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func (c *Container) ensureDockerImage(ctx context.Context, image string) error {
	// TODO: for now it just always pulls but we gotta make this smarter
	reader, err := c.client.ImagePull(ctx, image, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed image pull: %w", err)
	}

	defer func() { _ = reader.Close() }()
	_, _ = io.Copy(c.eventBus, reader)

	return nil
}

func (c *Container) Start(ctx context.Context) error {
	if !c.Exists() {
		return errors.New("container no no exist")
	}

	c.SetState("running")
	if _, err := c.client.ContainerStart(ctx, c.ID(), client.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed container start: %w", err)
	}

	return nil
}

func (c *Container) DebugCreate(ctx context.Context) error {
	if c.Exists() {
		return errors.New("container alr exist")
	}

	// TODO: this state system truly needs to be found a better solution for
	c.SetState("pulling")
	if err := c.ensureDockerImage(ctx, "docker.io/library/alpine"); err != nil {
		return fmt.Errorf("failed pulling image: %w", err)
	}

	resp, err := c.client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Cmd: []string{"echo", "hello, world\n"},
			Tty: false,
		},
		Image: "alpine",
	})
	if err != nil {
		return fmt.Errorf("failed container create: %w", err)
	}

	c.SetState("created")
	c.SetID(resp.ID)
	c.SetExists(true)

	return nil
}

func (c *Container) Attach(ctx context.Context) error {
	if !c.Exists() {
		return errors.New("container no no exist")
	}

	out, err := c.client.ContainerAttach(ctx, c.ID(), client.ContainerAttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	})
	if err != nil {
		return fmt.Errorf("failed container attach: %w", err)
	}

	go func() {
		defer out.Close()
		_, _ = io.Copy(c.eventBus, out.Reader)
	}()

	return nil
}

// TODO: Add Remove options
func (c *Container) Remove(ctx context.Context) error {
	if !c.Exists() {
		return errors.New("container no no exist")
	}

	if _, err := c.client.ContainerRemove(ctx, c.ID(), client.ContainerRemoveOptions{}); err != nil {
		return err
	}

	c.SetExists(false)

	return nil
}
