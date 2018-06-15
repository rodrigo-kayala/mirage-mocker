package main

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"net/http"
	"os"

	"github.com/miraeducation/mirage-mocker/config"
	"github.com/miraeducation/mirage-mocker/processor"
)

type miraFormatter struct{}

func (f *miraFormatter) Format(entry *log.Entry) ([]byte, error) {
	return []byte(
			fmt.Sprintf("%s %s %s\n",
				entry.Time.UTC().Format("2006-01-02 15:04:05.999"),
				strings.ToUpper(entry.Level.String()),
				entry.Message)),
		nil
}

var rp processor.Processor

func main() {
	configFile := "config.yml"
	if len(os.Args) > 1 {
		configFile = os.Args[1] + ".yml"
	}

	log.SetOutput(os.Stdout)
	log.SetFormatter(new(miraFormatter))

	c := config.LoadConfig(configFile)
	logLevel, err := log.ParseLevel(c.LogLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %s", c.LogLevel)
	}
	log.SetLevel(logLevel)

	log.Infof("Using config %s", configFile)

	rp, err := processor.NewFromConfig(c)

	if err != nil {
		log.Fatalf("Error creating processor: %v", err)
	}

	http.HandleFunc("/", rp.Process)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
