package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"rasanta/internal/app"
	"rasanta/internal/config"
)

func main() {
	// Konfigurasi dibaca sekali saat startup supaya dependency wiring tetap deterministic.
	cfg := config.Load()
	server := app.NewHTTPServer(cfg)

	// Jalankan server di goroutine supaya main goroutine bisa handle lifecycle app.
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Listen(":" + cfg.Port)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatal(err)
		}
	case sig := <-quit:
		log.Printf("signal diterima (%s), mulai shutdown server", sig.String())
	}

	if err := server.Shutdown(); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
