package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/numbereddev/zero-daemon/internal/git"
	apphttp "github.com/numbereddev/zero-daemon/internal/http"
	"github.com/numbereddev/zero-daemon/internal/server"
)

func main() {
	app := fiber.New()

	gitSvc := git.NewService()
	serverSvc := server.NewService()

	gitHdl := git.NewHandler(gitSvc)
	serviceHdl := server.NewHandler(serverSvc)

	apphttp.Register(app, apphttp.RouterDeps{
		GitHandler:    gitHdl,
		ServerHandler: serviceHdl,
	})

	log.Fatal(app.Listen(":3000"))
}
