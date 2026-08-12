package router

import (
	"github.com/gofiber/fiber/v3"
)

func Register(app *fiber.App) {
	{
		servers := app.Group("/servers")

		servers.Get("/", getServicesList)
		servers.Post("/", postCreateService)

		servers.Get("/:id", getService)
		servers.Post("/:id", postUpdateService)
		servers.Delete("/:id", deleteService)
	}

	{
		git := app.Group("/git", useGitSemaphore)
		repo := git.Group("/:repo")
		repo.Get("/info/refs", getGitInfoRefs)
		repo.Post("/git-upload-pack", postGitUploadPack)
		repo.Post("/git-receive-pack", postGitReceivePack)
	}
}
