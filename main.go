package main

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"

	"github.com/rodrigo-kayala/mirage-mocker/config"
	"github.com/rodrigo-kayala/mirage-mocker/processor"
)

// LoadConfig loads yaml configuration
func loadConfig(configPath string) config.Config {
	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatal().Err(err).Msg("error while reading config file")
	}

	m := config.Config{}
	err = yaml.Unmarshal(b, &m)

	if err != nil {
		log.Fatal().Err(err).Msg("error while unmarshalling yml")
	}

	return m
}

func main() {
	configFile := "config.yml"
	if len(os.Args) > 1 {
		configFile = os.Args[1] + ".yml"
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Output(os.Stdout)

	c := loadConfig(configFile)
	log.Info().Msgf("using config %s", configFile)

	rp, err := processor.NewFromConfig(c)

	if err != nil {
		log.Fatal().Err(err).Msg("error creating processor")
	}

	http.HandleFunc("/", rp.Process)
	log.Fatal().Err(http.ListenAndServe(":8080", nil)).Msg("error serving http")
}
