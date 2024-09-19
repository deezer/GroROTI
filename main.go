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

	"github.com/deezer/groroti/internal/middlewares"
	"github.com/deezer/groroti/internal/model"
	"github.com/deezer/groroti/internal/services"
	"github.com/deezer/groroti/internal/staticEmbed"
	"github.com/rs/zerolog/log"
)

var (
	Version string
	otelShutdown func(context.Context) error
)

func main() {
	if err := run(); err != nil {
		log.Fatal().Msgf(err.Error())
	}
}

func run() (err error) {
	// Get config first
	configRepository, err := services.GetConfig()
	if err != nil {
		return err
	}

	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err = middlewares.SetupOTelSDK(ctx, configRepository)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	services.Version = Version
	log.Info().Msgf("GroROTI version v%s", Version)
	sqliteDatabase := model.InitDatabase()
	defer sqliteDatabase.Close()

	// load embedded gotemplates in embed.Templates
	err = staticEmbed.LoadTemplates()
	if err != nil {
		return fmt.Errorf("couldn't load templates : %s", err.Error())
	}

	services.Register()
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