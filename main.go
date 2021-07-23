package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

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
	configFile := "mocker.yml"
	if envVarConfig := os.Getenv("MIRAGE_MOCKER_CONFIG"); envVarConfig != "" {
		configFile = envVarConfig
	} else {
		if len(os.Args) > 1 {
			configFile = os.Args[1]
		}
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	c := loadConfig(configFile)

	if c.PrettyLogs {
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		log.Logger = log.Logger.Output(output)
	}

	log.Info().Msgf("using config file: %s", configFile)
	log.Debug().Msgf("config content: %#v", c)

	rp, err := processor.NewFromConfig(c)

	if err != nil {
		log.Fatal().Err(err).Msg("error creating processor")
	}

	port := 8080
	if c.Port > 0 {
		port = c.Port
	}

	http.HandleFunc("/", rp.Process)
	log.Fatal().Err(http.ListenAndServe(fmt.Sprintf(":%d", port), nil)).Msg("error serving http")
}
