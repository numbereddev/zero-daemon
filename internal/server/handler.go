package server

import "github.com/gofiber/fiber/v3"

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Get("/", h.ListServers)
	router.Post("/", h.CreateServer)

	router.Get("/:id", func(c fiber.Ctx) {})
	router.Post("/:id", func(c fiber.Ctx) {})
	router.Delete("/:id", func(c fiber.Ctx) {})
}

func (h *Handler) ListServers(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

func (h *Handler) CreateServer(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}
