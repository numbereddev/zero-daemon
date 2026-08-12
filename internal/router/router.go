package router

import (
	"github.com/gofiber/fiber/v3"
)

func Register(app *fiber.App) {
	{
		servers := app.Group("/servers")

		servers.Get("/", getListServers)
		servers.Post("/", postCreateServer)

		servers.Get("/:id", getServer)
		servers.Post("/:id", postUpdateServer)
		servers.Delete("/:id", deleteServer)
	}

	{
		git := app.Group("/git", useGitSemaphore)
		repo := git.Group("/:repo")
		repo.Get("/info/refs", getGitInfoRefs)
		repo.Post("/git-upload-pack", postGitUploadPack)
		repo.Post("/git-receive-pack", postGitReceivePack)
	}
}
