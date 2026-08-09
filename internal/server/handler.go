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
	router.Get("/", func(c fiber.Ctx) {})
	router.Post("/", func(c fiber.Ctx) {})

	router.Get("/:id", func(c fiber.Ctx) {})
	router.Post("/:id", func(c fiber.Ctx) {})
	router.Delete("/:id", func(c fiber.Ctx) {})
}
