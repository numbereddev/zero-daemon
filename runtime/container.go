package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/containerd/errdefs"
	"github.com/moby/moby/client"
)

func (r *Runtime) Start(ctx context.Context) error {
	// TODO: first check here if already exists and running, if not then recreate, if yes resync state

	r.SetState("starting")

	if err := r.remove(ctx); err != nil {
		if !errdefs.IsNotFound(err) {
			r.SetState("offline")
			_ = r.bus.Publish(EventStateError, err.Error())
			return fmt.Errorf("failed container cleanup: %w", err)
		}
	}

	if err := r.create(ctx); err != nil {
		r.SetState("offline")
		_ = r.bus.Publish(EventStateError, err.Error())
		return fmt.Errorf("failed container creation: %w", err)
	}

	tCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := r.Attach(tCtx); err != nil {
		r.SetState("offline")
		_ = r.bus.Publish(EventStateError, err.Error())
		return fmt.Errorf("failed container attach: %w", err)
	}

	if _, err := r.cli.ContainerStart(tCtx, r.id, client.ContainerStartOptions{}); err != nil {
		r.SetState("offline")
		_ = r.bus.Publish(EventStateError, err.Error())
		return fmt.Errorf("failed container start: %w", err)
	}

	// TODO: implementing a running detector or something
	r.SetState("running")
	return nil
}

func (r *Runtime) Attach(ctx context.Context) error {
	if r.IsAttached() {
		return nil
	}

	if stream, err := r.cli.ContainerAttach(ctx, r.ID(), client.ContainerAttachOptions{
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

		writer := r.bus.Writer(EventConsoleOut)
		buf := make([]byte, 32*1024)

		for {
			n, err := r.stream.Reader.Read(buf)
			if n > 0 {
				_, _ = writer.Write(buf[:n])
			}

			if err != nil {
				return
			}
		}
	}()

	return nil
}

func (r *Runtime) Exists() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, err := r.cli.ContainerInspect(ctx, r.ID(), client.ContainerInspectOptions{}); err != nil {
		if errdefs.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
