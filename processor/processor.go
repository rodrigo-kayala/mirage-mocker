package processor

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/miraeducation/mirage-mocker/config"
)

// Processor structure
type Processor struct {
	Parsers []Parser
}

// Parser interface
type Parser interface {
	ProcessRequest(w http.ResponseWriter, r *http.Request)
	GetBaseParser() BaseParser
}

// Process current request and write response
func (rp Processor) Process(w http.ResponseWriter, r *http.Request) {
	requestProcess, err := rp.matchParser(r)
	log.Debugf("requestProcess: %v", requestProcess)

	if err != nil {
		errorResponse(w, fmt.Sprintf("Error processing request: %v", err), 500)
		return
	}
	requestProcess.ProcessRequest(w, r)
}

func (rp Processor) matchParser(r *http.Request) (Parser, error) {
	log.Debugf("Parsers: %v", rp.Parsers)
	for _, parser := range rp.Parsers {
		baseParser := parser.GetBaseParser()
		if baseParser.ContentType != "" && !strings.Contains(r.Header.Get("Content-Type"), baseParser.ContentType) {
			continue
		}

		if !containsMethod(baseParser.Methods, r.Method) {
			continue
		}

		match, err := regexp.MatchString(baseParser.Pattern, r.URL.Path)

		if err != nil {
			return nil, err
		}

		if match {
			return parser, nil
		}
	}
	return nil, errors.New("No match found for request")
}

// BaseParser base structure
type BaseParser struct {
	Pattern     string
	Methods     []string
	ContentType string
	Log         string
}

// NewFromConfig creates a new RequestProcessor from a Config struct
func NewFromConfig(c config.Config) (rp Processor, err error) {
	rp = Processor{}
	for _, service := range c.Services {
		baseParser := BaseParser{
			ContentType: service.Parser.ContentType,
			Log:         service.Parser.Log,
			Methods:     service.Parser.Methods,
			Pattern:     service.Parser.Pattern,
		}

		switch service.Parser.ConfigType {
		case "pass":
			rp.Parsers = append(rp.Parsers, createPassParser(baseParser, service.Parser))
		case "mock":
			mockParser := MockParser{}
			mockParser.BaseParser = baseParser
			baseResp := BaseResponse{Status: service.Parser.Response.Status}

			switch service.Parser.Response.BodyType {

			case "fixed":
				mockParser.Response = ResponseFixed{
					BaseResponse: baseResp,
					Body:         service.Parser.Response.Body,
					ContentType:  service.Parser.Response.ContentType,
				}
				rp.Parsers = append(rp.Parsers, mockParser)
			case "request":
				mockParser.Response = ResponseRequest{
					BaseResponse: baseResp,
					ContentType:  service.Parser.Response.ContentType,
				}
				rp.Parsers = append(rp.Parsers, mockParser)
			case "runnable":
				runnable, err := runnableMethod(
					service.Parser.Response.ResponseLib,
					service.Parser.Response.ResponseSymbol,
				)

				if err != nil {
					log.Fatalf("Error processing transform method %v", err)
				}

				mockParser.Response = ResponseRunnable{
					BaseResponse: baseResp,
					runnable:     runnable,
				}
				rp.Parsers = append(rp.Parsers, mockParser)
			default:
				log.Fatalf("Bad value for body-type %s", service.Parser.Response.BodyType)
			}

		default:
			log.Fatalf("Bad value for config-type %s", service.Parser.ConfigType)
		}

	}

	return rp, nil
}

func containsMethod(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
