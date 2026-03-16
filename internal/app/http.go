package app

import (
	"rasanta/internal/config"
	"rasanta/internal/handler"
	"rasanta/internal/repository"
	"rasanta/internal/router"
	"rasanta/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func NewHTTPServer(cfg config.Config) *fiber.App {
	// Initialisasi semua dependency
	repo := repository.NewGoogleMapsRepository(cfg.HTTPTimeout, cfg.GoogleMapsUserAgent, cfg.GoogleMapsCookie)
	placeService := service.NewPlaceService(repo)
	placeHandler := handler.NewPlaceHandler(placeService)

	app := fiber.New(fiber.Config{AppName: cfg.AppName})

	// Tambahkan middleware CORS untuk mengizinkan permintaan dari frontend
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "*",
		AllowMethods: "*",
	}))

	// Register semua route
	router.Register(app, placeHandler)

	return app
}
