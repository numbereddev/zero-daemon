package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/numbereddev/zero-daemon/internal/db"
	"github.com/numbereddev/zero-daemon/internal/router"
)

func main() {
	_ = db.Init()
	if err := db.Migrate(); err != nil {
		panic(fmt.Errorf("could not migrate: %v", err))
	}

	app := fiber.New()
	router.Register(app)
	log.Fatal(app.Listen(":3000"))
}
