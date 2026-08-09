package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/numbereddev/zero-daemon/internal/git"
	"github.com/numbereddev/zero-daemon/internal/service"
)

type RouterDeps struct {
	GitHandler     *git.Handler
	ServiceHandler *service.Handler
}

func Register(app *fiber.App, deps RouterDeps) {
	deps.GitHandler.RegisterRoutes(app.Group("/git"))
	deps.ServiceHandler.RegisterRoutes(app.Group("/services"))
}
