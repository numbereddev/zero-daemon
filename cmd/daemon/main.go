package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/sse"
	"github.com/moby/moby/client"
	"github.com/numbereddev/zero-daemon/internal/database"
	"github.com/numbereddev/zero-daemon/router"
	"github.com/numbereddev/zero-daemon/service"
)

func main() {
	_ = database.Init()
	if err := database.Migrate(); err != nil {
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
	debug.Get("/container/test", sse.New(sse.Config{
		Handler: func(c fiber.Ctx, stream *sse.Stream) error {
			if err := stream.Comment("connected"); err != nil {
				return err
			}

			runtime := service.New(apiClient, "test")
			events := runtime.Events()
			logs := events.On()
			defer events.Off(logs)

			errCh := make(chan error)
			go func() {
				defer close(errCh)
				ctx := context.Background()

				if err := runtime.Create(ctx); err != nil {
					errCh <- fmt.Errorf("failed creating container: %w", err)
					return
				}

				if err := runtime.Start(ctx); err != nil {
					errCh <- fmt.Errorf("failed starting container: %w", err)
					return
				}

				if err := runtime.Remove(ctx); err != nil {
					errCh <- fmt.Errorf("failed removing bontainer: %w", err)
					return
				}
			}()

			for {
				select {
				// logs
				case line, ok := <-logs:
					if !ok {
						return nil
					}

					if err := stream.Event(sse.Event{
						Name: "Received",
						Data: string(line),
					}); err != nil {
						return err
					}
				// errors
				case err, ok := <-errCh:
					if ok {
						fmt.Printf("error occurred: %v\n", err)
						return stream.Event(sse.Event{
							Name: "Error",
							Data: err.Error(),
						})
					}
				// done
				case <-c.Context().Done():
					return nil
				}
			}
		},
	}))

	log.Fatal(app.Listen(":3000"))
}
