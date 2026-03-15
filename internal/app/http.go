package app

import (
	"rasanta/internal/config"
	"rasanta/internal/handler"
	"rasanta/internal/repository"
	"rasanta/internal/router"
	"rasanta/internal/service"

	"github.com/gofiber/fiber/v2"
)

func NewHTTPServer(cfg config.Config) *fiber.App {
	// Initialisasi semua dependency
	repo := repository.NewGoogleMapsRepository(cfg.HTTPTimeout, cfg.GoogleMapsUserAgent, cfg.GoogleMapsCookie)
	placeService := service.NewPlaceService(repo)
	placeHandler := handler.NewPlaceHandler(placeService)

	app := fiber.New(fiber.Config{AppName: cfg.AppName})
	// Register semua route
	router.Register(app, placeHandler)

	return app
}
