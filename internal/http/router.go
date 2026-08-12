package http

import (
	"github.com/gofiber/fiber/v3"
	"github.com/numbereddev/zero-daemon/internal/http/git"
	"github.com/numbereddev/zero-daemon/internal/http/server"
)

type RouterDeps struct {
	GitHandler    *git.Handler
	ServerHandler *server.Handler
}

func Register(app *fiber.App, deps RouterDeps) {
	deps.GitHandler.RegisterRoutes(app.Group("/git"))
	deps.ServerHandler.RegisterRoutes(app.Group("/services"))
}
