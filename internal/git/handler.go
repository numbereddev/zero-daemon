package git

import "github.com/gofiber/fiber/v3"

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

// gitSem prevents too many process forks from happening at once and overloading the daemon/server
var gitSem = make(chan struct{}, 100)

func (h *Handler) RegisterRoutes(router fiber.Router) {
	gitURL := router.Group("/:repo")

	// TODO: Support for git submodules (pulling from them + private repositories, will need handling).
	// For now: disabling submodule support.

	gitURL.Post("/git-upload-pack", h.UploadPack)
	gitURL.Post("/git-receive-pack", h.ReceivePack)
	gitURL.Get("/info/refs", h.InfoRefs)
}

func (h *Handler) UploadPack(c fiber.Ctx) error {
	gitSem <- struct{}{}
	defer func() { <-gitSem }()

	return c.SendStatus(fiber.StatusNotImplemented)
}

func (h *Handler) ReceivePack(c fiber.Ctx) error {
	gitSem <- struct{}{}
	defer func() { <-gitSem }()

	return c.SendStatus(fiber.StatusNotImplemented)
}

func (h *Handler) InfoRefs(c fiber.Ctx) error {
	gitSem <- struct{}{}
	defer func() { <-gitSem }()

	// service may be one of the two
	// - git-upload-pack ; client wants to fetch/clone
	// - git-receive-pack ; client wants to push
	// if it's not provided, it's a dumb http client (2009)
	// and will need to be rejected
	repoName := c.Params("repo")
	service := c.Query("service")
	if service != "git-upload-pack" && service != "git-receive-pack" {
		// Filter out Git dumb HTTP clients
		return c.SendStatus(fiber.StatusBadRequest)
	}

	_ = repoName

	// Git smart HTTP
	return c.SendStatus(fiber.StatusNotImplemented)
}
