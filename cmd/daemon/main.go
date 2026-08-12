package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/sse"
	"github.com/moby/moby/client"
	"github.com/numbereddev/zero-daemon/internal/container"
	"github.com/numbereddev/zero-daemon/internal/db"
	"github.com/numbereddev/zero-daemon/internal/router"
)

func main() {
	_ = db.Init()
	if err := db.Migrate(); err != nil {
		panic(fmt.Errorf("could not migrate: %v", err))
	}

	// Moby Docker client
	apiClient, err := client.New(client.FromEnv, client.WithUserAgent("zero-daemon/1.0.0"))
	if err != nil {
		log.Fatalf("failed moby client: %v\n", err)
	}
	defer func() { _ = apiClient.Close() }()

	app := fiber.New()
	app.Use(logger.New())
	router.Register(app)

	debug := app.Group("/debug")
	debug.Post("/container/create", func(c fiber.Ctx) error {
		// ctx := context.Background()

		return nil
	})

	debug.Get("/container/test", sse.New(sse.Config{
		Handler: func(c fiber.Ctx, stream *sse.Stream) error {
			container := container.Container{Client: apiClient}

			logsChan := container.Subscribe()
			defer container.Unsubscribe(logsChan)

			errCh := make(chan error)
			defer close(errCh)
			go func() {
				ctx := context.Background()

				if err := container.DebugPull(ctx); err != nil {
					errCh <- fmt.Errorf("failed pullng image: %w", err)
					return
				}

				if err := container.DebugCreate(ctx); err != nil {
					errCh <- fmt.Errorf("failed creating container: %w", err)
					return
				}

				if err := container.DebugAttach(ctx); err != nil {
					errCh <- fmt.Errorf("failed attaching to container: %w", err)
					return
				}

				if err := container.DebugStart(ctx); err != nil {
					errCh <- fmt.Errorf("failed starting container: %w", err)
					return
				}

				if err := container.DebugRemove(ctx); err != nil {
					errCh <- fmt.Errorf("failed removing bontainer: %w", err)
					return
				}
			}()

			for {
				select {
				case <-c.Context().Done():
					return nil
				case err := <-errCh:
					fmt.Printf("error occurred: %v\n", err)
					if err := stream.Event(sse.Event{
						Name: "Error",
						Data: err.Error(),
					}); err != nil {
						return err
					}
					return err
				case msg, ok := <-logsChan:
					if !ok {
						return nil
					}

					if err := stream.Event(sse.Event{
						Name: "Received",
						Data: string(msg),
					}); err != nil {
						return err
					}
				}
			}
		},
	}))

	log.Fatal(app.Listen(":3000"))
}
