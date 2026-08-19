package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/buger/jsonparser"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

// create creates the underlying Docker container and shouldn't be called manually
func (r *Server) create(ctx context.Context) error {
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
func (r *Server) remove(ctx context.Context) error {
	if _, err := r.cli.ContainerRemove(ctx, r.ID(), client.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		return err
	}

	return nil
}

type ImagePullProgress struct {
	Status  string
	Current uint64
	Total   uint64
}

// checkImage checks an image and potentially pulls it, it's a non-blocking function and outputs
// the status updates like docker log lines or errors through the channel
func (r *Server) checkImage(image string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// TODO: for now it just always pulls but we gotta make this smarter

	r.bus.Publish(EventImagePullBegin, fmt.Sprintf("pulling image `%s`...", image))
	out, err := r.cli.ImagePull(ctx, image, client.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer out.Close()

	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Bytes()

		status, _ := jsonparser.GetString(line, "status")
		current, _ := jsonparser.GetUint64(line, "progressDetail", "current")
		total, _ := jsonparser.GetUint64(line, "progressDetail", "total")

		if payload, err := json.Marshal(ImagePullProgress{
			Status:  status,
			Current: current,
			Total:   total,
		}); err != nil {
			return err
		} else {
			r.bus.Publish(EventImagePullProgress, string(payload))
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	r.bus.Publish(EventImagePullDone, fmt.Sprintf("finished pulling image `%s`.", image))
	return nil
}

func (r *Server) setStream(stream *client.HijackedResponse) {
	r.mu.Lock()
	r.stream = stream
	r.mu.Unlock()
}
