package config

// Config configuration yaml structure
type Config struct {
	Port       int       `yaml:"port"`
	PrettyLogs bool      `yaml:"pretty-logs"`
	Services   []Service `yaml:"services"`
}

// Service yaml structure
type Service struct {
	Parser Parser `yaml:"parser"`
}

// Parser yaml structure
type Parser struct {
	Pattern         string            `yaml:"pattern"`
	Rewrites        []Rewrite         `yaml:"rewrite"`
	Methods         []string          `yaml:"methods"`
	Headers         map[string]string `yaml:"headers"`
	ConfigType      string            `yaml:"type"`
	TransformLib    string            `yaml:"transform-lib"`
	TransformSymbol string            `yaml:"transform-symbol"`
	Responses       []Response        `yaml:"responses"`
	PassBaseURI     string            `yaml:"pass-base-uri"`
	Log             bool              `yaml:"log"`
}

// Rewrite yaml structure
type Rewrite struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
}

// Response yaml structure
type Response struct {
	Headers           map[string]string `yaml:"headers"`
	Status            map[string]int    `yaml:"status"`
	BodyType          string            `yaml:"body-type"`
	Body              string            `yaml:"body"`
	BodyFile          string            `yaml:"body-file"`
	ResponseLib       string            `yaml:"response-lib"`
	ResponseSymbol    string            `yaml:"response-symbol"`
	MagicHeaderName   string            `yaml:"magic-header-name"`
	MagicHeaderFolder string            `yaml:"magic-header-folder"`
	Distribuition     float64           `yaml:"distribution"`
	Delay             Delay             `yaml:"delay"`
}

type Delay struct {
	Min string `yaml:"min"`
	Max string `yaml:"max"`
}
