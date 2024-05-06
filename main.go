package main

import (
	"github.com/deezer/groroti/internal/model"
	"github.com/deezer/groroti/internal/server"
	"github.com/deezer/groroti/internal/services"
	"github.com/deezer/groroti/internal/staticEmbed"
	"github.com/rs/zerolog/log"
)

var Version string

func main() {
	log.Info().Msgf("GroROTI version v%s", Version)
	sqliteDatabase := model.InitDatabase()
	defer sqliteDatabase.Close()

	// load embedded gotemplates in embed.Templates
	err := staticEmbed.LoadTemplates()
	if err != nil {
		log.Fatal().Msgf("Couldn't load templates : %s", err.Error())
	}

	services.Version = Version
	services.Register()
	if err := server.Start(); err != nil {
		log.Fatal().AnErr("The server couldn't start", err)
	}

}
