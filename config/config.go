package config

import (
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Config configuration yaml structure
type Config struct {
	LogLevel string    `yaml:"log-level"`
	Services []Service `yaml:"services"`
}

// Service yaml structure
type Service struct {
	Parser Parser `yaml:"parser"`
}

// Parser yaml structure
type Parser struct {
	Pattern         string    `yaml:"pattern"`
	Rewrites        []Rewrite `yaml:"rewrite"`
	Methods         []string  `yaml:"methods"`
	ContentType     string    `yaml:"content-type"`
	ConfigType      string    `yaml:"type"`
	TransformLib    string    `yaml:"transform-lib"`
	TransformSymbol string    `yaml:"transform-symbol"`
	Response        Response  `yaml:"response"`
	PassBaseURI     string    `yaml:"pass-base-uri"`
	Log             string    `yaml:"log"`
}

// Rewrite yaml structure
type Rewrite struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

// Response yaml structure
type Response struct {
	ContentType    string         `yaml:"content-type"`
	Status         map[string]int `yaml:"status"`
	BodyType       string         `yaml:"body-type"`
	Body           string         `yaml:"body"`
	ResponseLib    string         `yaml:"response-lib"`
	ResponseSymbol string         `yaml:"response-symbol"`
}

// LoadConfig loads yaml configuration
func LoadConfig(configPath string) Config {
	b, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	m := Config{}
	err = yaml.Unmarshal(b, &m)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return m
}
