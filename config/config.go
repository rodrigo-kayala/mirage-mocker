package config

// Config configuration yaml structure
type Config struct {
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
	Log             bool      `yaml:"log"`
	Delay           Delay     `yaml:"delay"`
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
	BodyFile       string         `yaml:"body-file"`
	ResponseLib    string         `yaml:"response-lib"`
	ResponseSymbol string         `yaml:"response-symbol"`
}

type Delay struct {
	Min string `yaml:"min"`
	Max string `yaml:"max"`
}
