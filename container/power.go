package container

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
)

func (c *Container) Start(ctx context.Context) error {
	// TODO: first check here if already exists and running, if not then recreate, if yes resync state
	c.setStatus("starting")

	if err := c.Remove(ctx); err != nil {
		if !errdefs.IsNotFound(err) {
			c.setStatus("offline")
			c.bus.Publish(EventStateError, err.Error())
			return fmt.Errorf("failed container cleanup: %w", err)
		}
	}

	if err := c.Create(ctx); err != nil {
		c.setStatus("offline")
		c.bus.Publish(EventStateError, err.Error())
		return fmt.Errorf("failed container creation: %w", err)
	}

	sctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := c.Attach(sctx); err != nil {
		c.setStatus("offline")
		c.bus.Publish(EventStateError, err.Error())
		return fmt.Errorf("failed container attach: %w", err)
	}

	if _, err := c.cli.ContainerStart(sctx, c.id, client.ContainerStartOptions{}); err != nil {
		c.setStatus("offline")
		c.bus.Publish(EventStateError, err.Error())
		return fmt.Errorf("failed container start: %w", err)
	}

	// TODO: implementing a running detector or something
	c.setStatus("running")
	return nil
}

func (c *Container) Stop(ctx context.Context) error {
	// TODO: stop options
	c.setStatus("stopping")

	// TODO: make stop commands and stuff exist, too. this is just temporary
	if _, err := c.cli.ContainerStop(ctx, c.ID(), client.ContainerStopOptions{
		Timeout: new(60),
	}); err != nil && !errdefs.IsNotFound(err) {
		return err
	}

	if err := c.Remove(ctx); err != nil && !errdefs.IsNotFound(err) {
		return err
	}

	c.setStatus("offline")
	return nil
}

func (c *Container) Restart(ctx context.Context) error {
	if err := c.Stop(ctx); err != nil {
		return err
	}

	if err := c.Start(ctx); err != nil {
		return err
	}

	return nil
}

func (c *Container) Status() string {
	return c.status.Load()
}

func (c *Container) setStatus(state string) {
	if c.Status() == state {
		return
	}

	c.status.Store(state)
	c.bus.Publish(EventStateChange, state)
}
