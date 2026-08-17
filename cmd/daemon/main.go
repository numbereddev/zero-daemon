package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/moby/moby/client"
	"github.com/numbereddev/zero-daemon/events"
	"github.com/numbereddev/zero-daemon/internal/database"
	"github.com/numbereddev/zero-daemon/router"
	"github.com/numbereddev/zero-daemon/runtime"
)

type WSMessage struct {
	Event string            `json:"event"`
	Args  []json.RawMessage `json:"args"`
}

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
	app.Use("/ws", func(c fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	router.Register(app)

	{
		// WARNING: TESTING STUFF

		testRun := runtime.New(apiClient, "test")

		app.Get("/testing", func(c fiber.Ctx) error {
			ctx := context.Background()

			if err := testRun.Start(ctx); err != nil {
				return c.
					Status(fiber.StatusInternalServerError).
					SendString(fmt.Sprintf("failed starting container: %v", err))
			}

			return nil
		})

		app.Get("/ws/testing/console", websocket.New(func(c *websocket.Conn) {
			_ = testRun.Attach(context.Background())

			eventsCh := testRun.Events().On([]string{events.TopicAll}, 250)
			defer testRun.Events().Off(eventsCh)

			disconnect := make(chan struct{})
			defer func() {
				close(disconnect)
				_ = c.Close()
			}()

			// Container stdout -> Websocket
			// TODO: extract/abstract the batcher
			go func() {
				ticker := time.NewTicker(30 * time.Millisecond)
				defer ticker.Stop()
				var batchBuffer bytes.Buffer

				for {
					select {
					case <-disconnect:
						return

					case event, ok := <-eventsCh:
						if !ok {
							return
						}

						if event.Topic == runtime.EventConsoleOut {
							batchBuffer.WriteString(event.Data)
						} else {
							payload, err := json.Marshal(WSMessage{
								Event: event.Topic,
								Args:  []json.RawMessage{mustJSONString(event.Data)},
							})
							if err == nil {
								_ = c.WriteMessage(websocket.TextMessage, payload)
							}
						}

					case <-ticker.C:
						if batchBuffer.Len() > 0 {
							payload, err := json.Marshal(WSMessage{
								Event: runtime.EventConsoleOut,
								Args:  []json.RawMessage{mustJSONString(batchBuffer.String())},
							})

							if err == nil {
								if err := c.WriteMessage(websocket.TextMessage, payload); err != nil {
									return // Client disconnected
								}
							}

							batchBuffer.Reset() // empty buffer for next 30ms
						}
					}
				}
			}()

			// Websocket -> Container stdin (xTerm)
			for {
				_, rawMsg, err := c.ReadMessage()
				if err != nil {
					// Client disconnected
					break
				}

				var msg WSMessage
				if err := json.Unmarshal(rawMsg, &msg); err != nil {
					continue
				}

				fmt.Printf("Message: %+v\n", msg)

				// TODO: send sdtin
				// if msg.Event == "send command" && len(msg.Args) > 0 {
				// 	var cmd string
				// 	if err := json.Unmarshal(msg.Args[0], &cmd); err == nil {
				// 		svc.SendStdin([]byte(cmd))
				// 	}
				// }
			}
		}))
	}

	log.Fatal(app.Listen(":3000"))
}

func mustJSONString(s string) []byte {
	b, _ := json.Marshal(s)
	return b
}
