package router

import "github.com/gofiber/fiber/v3"

var gitSemaphore = make(chan struct{}, 100)

func useGitSemaphore(c fiber.Ctx) error {
	gitSemaphore <- struct{}{}
	defer func() { <-gitSemaphore }()
	return c.Next()
}

func postGitUploadPack(c fiber.Ctx) error {
	// TODO: git handling
	return c.SendStatus(fiber.StatusNotImplemented)
}

func postGitReceivePack(c fiber.Ctx) error {
	// TODO: git handling
	return c.SendStatus(fiber.StatusNotImplemented)
}

func getGitInfoRefs(c fiber.Ctx) error {
	// service may be one of these two
	// - git-upload-pack ; client wants to fetch/clone
	// - git-receive-pack ; client wants to push
	// if it's not provided, it's a dumb http client (from 2009)
	// and will need to be rejected
	repoName := c.Params("repo")
	service := c.Query("service")
	if service != "git-upload-pack" && service != "git-receive-pack" {
		return c.SendStatus(fiber.StatusNotFound)
	}

	_ = repoName

	// TODO: git handling

	return c.SendStatus(fiber.StatusNotImplemented)
}
