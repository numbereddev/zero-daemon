package runtime

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

// create creates the underlying Docker container and shouldn't be called manually
func (r *Runtime) create(ctx context.Context) error {
	// TODO: proper service configuration stuff
	image := "docker.io/library/ubuntudksfjaskf"
	if err := r.checkImage(image); err != nil {
		return fmt.Errorf("failed pulling image: %w", err)
	}

	const defaultPeriod int64 = 100_000
	const cpuPercentage int64 = 50
	cpuQuota := (cpuPercentage * defaultPeriod) / 100

	_, err := r.cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		// This is okay, because Docker lets us do everything on containers with both the Docker ID
		// or the container name interchangeably.
		Name: r.ID(),
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

// TODO: Add remove options
func (r *Runtime) remove(ctx context.Context) error {
	if _, err := r.cli.ContainerRemove(ctx, r.ID(), client.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		return err
	}

	return nil
}

func (r *Runtime) checkImage(image string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// TODO: for now it just always pulls but we gotta make this smarter

	_ = r.bus.Publish(EventImagePullBegin, fmt.Sprintf("pulling image `%s`...", image))
	reader, err := r.cli.ImagePull(ctx, image, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed image pull: %w", err)
	}

	defer func() { _ = reader.Close() }()
	// TODO: make it actually parse the lines and send it properly in
	_, _ = io.Copy(r.bus.Writer(EventImagePullProgress), reader)
	_ = r.bus.Publish(EventImagePullDone, fmt.Sprintf("finished pulling image `%s`.", image))

	return nil
}

func (r *Runtime) setStream(stream *client.HijackedResponse) {
	r.mu.Lock()
	r.stream = stream
	r.mu.Unlock()
}
