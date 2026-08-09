package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/numbereddev/zero-daemon/internal/git"
	apphttp "github.com/numbereddev/zero-daemon/internal/http"
	"github.com/numbereddev/zero-daemon/internal/service"
)

func main() {
	app := fiber.New()

	gitSvc := git.NewService()
	serviceSvc := service.NewService()

	gitHdl := git.NewHandler(gitSvc)
	serviceHdl := service.NewHandler(serviceSvc)

	apphttp.Register(app, apphttp.RouterDeps{
		GitHandler:     gitHdl,
		ServiceHandler: serviceHdl,
	})

	log.Fatal(app.Listen(":3000"))
}
