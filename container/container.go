package container

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	"github.com/numbereddev/zero-daemon/events"
	"github.com/numbereddev/zero-daemon/internal/system"
)

const (
	EventImagePullProgress = "image.pull.progress" // event.Data is of type ImagePullProgress
	EventImagePullBegin    = "image.pull.begin"
	EventImagePullDone     = "image.pull.done"
	EventStateChange       = "state.change"
	EventStateError        = "state.error"
	EventConsoleOut        = "console.output"
)

// TODO: crash detector
// TODO: start/running detector
// TODO: historical logs getting

type ImagePullProgress struct {
	Status  string
	Current uint64
	Total   uint64
}

type Container struct {
	mu  sync.RWMutex
	cli *client.Client

	id     string
	status system.AtomicString
	config *Config

	stream *client.HijackedResponse
	bus    *events.Bus[string]
}

func New(cli *client.Client, id string, config *Config) *Container {
	if config == nil {
		panic("config is required")
	}

	// TODO: Configuration
	return &Container{
		id:     id,
		cli:    cli,
		config: config,
		status: *system.NewAtomicString("offline"),
		bus:    events.NewBus[string](),
	}
}

func (c *Container) ID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.id
}

func (c *Container) Events() events.Reader[string] {
	return c.bus
}

func (c *Container) Config() *Config {
	return c.config
}

func (c *Container) IsAttached() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stream != nil
}

func (c *Container) setStream(stream *client.HijackedResponse) {
	c.mu.Lock()
	c.stream = stream
	c.mu.Unlock()
}

// Attach attaches the stream to the container events and sets up resource polling
func (c *Container) Attach(ctx context.Context) error {
	if c.IsAttached() {
		return nil
	}

	if stream, err := c.cli.ContainerAttach(ctx, c.ID(), client.ContainerAttachOptions{
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Stream: true,
	}); err != nil {
		return fmt.Errorf("failed container attach: %w", err)
	} else {
		c.setStream(&stream.HijackedResponse)
	}

	go func() {
		defer func() {
			c.stream.Close()
			c.setStream(nil)
		}()

		writer := c.bus.Writer(EventConsoleOut)
		buf := make([]byte, 32*1024)
		for {
			n, err := c.stream.Reader.Read(buf)
			if n > 0 {
				writer.Write(buf[:n])
			}
			if err != nil {
				return
			}
		}
	}()

	return nil
}

// Create creates the underlying Docker container
func (c *Container) Create(ctx context.Context) error {
	// TODO: proper service configuration stuff
	image := "docker.io/library/ubuntu"
	if err := c.checkImage(image); err != nil {
		return fmt.Errorf("failed pulling image: %w", err)
	}

	const defaultPeriod int64 = 100_000
	const cpuPercentage int64 = 50
	cpuQuota := (cpuPercentage * defaultPeriod) / 100

	_, err := c.cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		// This is okay, because Docker lets us do everything on containers with both the Docker ID
		// or the container name interchangeably.
		Name: c.ID(),
		HostConfig: &container.HostConfig{
			Resources: container.Resources{
				CPUPeriod:  defaultPeriod,
				CPUQuota:   cpuQuota,
				Memory:     1024 * 1024 * 1024,
				MemorySwap: 1024 * 1024 * 1024,
				// BlkioWeight: 512,
			},
		},
		Config: &container.Config{
			Cmd: []string{"sh", "-c", `apt update -y && apt upgrade -y && apt install htop -y && htop`},

			Tty: true,
		},
		Image: image,
	})
	if err != nil {
		return fmt.Errorf("failed container create: %w", err)
	}

	return nil
}

func (c *Container) Remove(ctx context.Context) error {
	// TODO: Add Remove options
	if _, err := c.cli.ContainerRemove(ctx, c.ID(), client.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		return err
	}

	return nil
}

func (c *Container) Exists() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := c.cli.ContainerInspect(ctx, c.ID(), client.ContainerInspectOptions{}); err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
