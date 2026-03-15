package router

import (
	_ "embed"
	"rasanta/internal/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func Register(app *fiber.App, placeHandler *handler.PlaceHandler) {
	app.Use(recover.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	api := app.Group("/api/v1")

	// Places
	api.Get("/places", placeHandler.Fetch)
	api.Get("/places/detail", placeHandler.Detail)
}
