package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/deezer/groroti/internal/model"
	"github.com/deezer/groroti/internal/services"
	"github.com/deezer/groroti/internal/staticEmbed"
	"github.com/rs/zerolog/log"
)

var Version string

func main() {
	if err := run(); err != nil {
		log.Fatal().Msgf(err.Error())
	}
}

func run() (err error) {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err := services.SetupOTelSDK(ctx)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	log.Info().Msgf("GroROTI version v%s", Version)
	sqliteDatabase := model.InitDatabase()
	defer sqliteDatabase.Close()

	// load embedded gotemplates in embed.Templates
	err = staticEmbed.LoadTemplates()
	if err != nil {
		return fmt.Errorf("Couldn't load templates : %s", err.Error())
	}

	services.Version = Version
	services.Register()

	configRepository, err := services.GetConfig()
	if err != nil {
		return err
	}
	addr := configRepository.BuildServerAddr()

	log.Info().Msgf("Start listening on %s", addr)
	http.ListenAndServe(addr, nil)

	// Start HTTP server.
	srv := &http.Server{
		Addr:         addr,
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
	}
	srvErr := make(chan error, 1)

	go func() {
		srvErr <- srv.ListenAndServe()
	}()


	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = srv.Shutdown(context.Background())
	return
}