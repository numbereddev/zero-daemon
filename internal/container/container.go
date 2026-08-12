package container

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type Container struct {
	id     string
	Client *client.Client

	// TODO: extract to subscription/sink system
	subsMx sync.RWMutex
	subs   map[chan []byte]bool
}

func (c *Container) DebugPull(ctx context.Context) error {
	reader, err := c.Client.ImagePull(ctx, "docker.io/library/alpine", client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed image pull: %w", err)
	}
	defer func() { _ = reader.Close() }()
	io.Copy(c, reader)
	return nil
}

func (c *Container) DebugCreate(ctx context.Context) error {
	resp, err := c.Client.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config: &container.Config{
			Cmd: []string{"echo", "hello, world\n"},
			Tty: false,
		},
		Image: "alpine",
	})
	if err != nil {
		return fmt.Errorf("failed container create: %w", err)
	}
	c.id = resp.ID

	return nil
}

func (c *Container) DebugStart(ctx context.Context) error {
	if c.id == "" {
		return errors.New("container no no exist")
	}

	if _, err := c.Client.ContainerStart(ctx, c.id, client.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("failed container start: %w", err)
	}

	return nil
}

func (c *Container) DebugAttach(ctx context.Context) error {
	if c.id == "" {
		return errors.New("container no no exist")
	}

	out, err := c.Client.ContainerAttach(ctx, c.id, client.ContainerAttachOptions{
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
		io.Copy(c, out.Reader)
	}()

	return nil
}

func (c *Container) DebugRemove(ctx context.Context) error {
	if c.id == "" {
		return errors.New("container no no exist")
	}

	if _, err := c.Client.ContainerRemove(ctx, c.id, client.ContainerRemoveOptions{}); err != nil {
		return err
	}

	return nil
}

// TODO: extract to subscription/sink system
func (c *Container) Subscribe() chan []byte {
	c.subsMx.Lock()
	defer c.subsMx.Unlock()

	if c.subs == nil {
		c.subs = make(map[chan []byte]bool)
	}

	ch := make(chan []byte, 10)
	c.subs[ch] = true
	return ch
}

func (c *Container) Unsubscribe(ch chan []byte) {
	c.subsMx.Lock()
	defer c.subsMx.Unlock()

	delete(c.subs, ch)
	close(ch)
}

func (c *Container) Broadcast(data []byte) {
	c.subsMx.RLock()
	defer c.subsMx.RUnlock()
	for sub := range c.subs {
		buf := make([]byte, len(data))
		copy(buf, data)

		select {
		case sub <- buf:
		default:
		}
	}
}

func (c *Container) Write(p []byte) (n int, err error) {
	c.Broadcast(p)
	return len(p), nil
}
