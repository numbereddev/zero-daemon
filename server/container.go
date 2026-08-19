package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/buger/jsonparser"
	"github.com/containerd/errdefs"
	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

func (s *Server) Attach(ctx context.Context) error {
	if s.IsAttached() {
		return nil
	}

	if stream, err := s.cli.ContainerAttach(ctx, s.ID(), client.ContainerAttachOptions{
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
		defer func() {
			s.stream.Close()
			s.setStream(nil)
		}()

		writer := s.bus.Writer(EventConsoleOut)
		buf := make([]byte, 32*1024)
		for {
			n, err := s.stream.Reader.Read(buf)
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
func (s *Server) Create(ctx context.Context) error {
	// TODO: proper service configuration stuff
	image := "docker.io/library/ubuntu"
	if err := s.checkImage(image); err != nil {
		return fmt.Errorf("failed pulling image: %w", err)
	}

	const defaultPeriod int64 = 100_000
	const cpuPercentage int64 = 50
	cpuQuota := (cpuPercentage * defaultPeriod) / 100

	_, err := s.cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		// This is okay, because Docker lets us do everything on containers with both the Docker ID
		// or the container name interchangeably.
		Name: s.ID(),
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

// checkImage checks an image and potentially pulls it, it's a non-blocking function and outputs
// the status updates like docker log lines or errors through the channel
func (s *Server) checkImage(image string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	// TODO: for now it just always pulls but we gotta make this smarter

	s.bus.Publish(EventImagePullBegin, fmt.Sprintf("pulling image `%s`...", image))
	out, err := s.cli.ImagePull(ctx, image, client.ImagePullOptions{})
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
			s.bus.Publish(EventImagePullProgress, string(payload))
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	s.bus.Publish(EventImagePullDone, fmt.Sprintf("finished pulling image `%s`.", image))
	return nil
}

func (s *Server) Exists() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := s.cli.ContainerInspect(ctx, s.ID(), client.ContainerInspectOptions{}); err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *Server) Remove(ctx context.Context) error {
	// TODO: Add Remove options
	if _, err := s.cli.ContainerRemove(ctx, s.ID(), client.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		return err
	}

	return nil
}

func (s *Server) setStream(stream *client.HijackedResponse) {
	s.mu.Lock()
	s.stream = stream
	s.mu.Unlock()
}
