package server

import (
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/deezer/groroti/internal/services"
)

func Start() error {
	configRepository, err := services.GetConfig()
	if err != nil {
		return err
	}

	addr := configRepository.BuildServerAddr()

	log.Info().Msgf("Start listening on %s", addr)
	return http.ListenAndServe(addr, nil)
}
