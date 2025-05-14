package shared

import (
	_ "embed"

	"github.com/charmbracelet/log"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

var (
	// Config manager
	Koanf *koanf.Koanf
)

//go:embed config/default.json
var defaultConfig []byte

func LoadUserConfig(path string) {
	log.Debug("Loading configuration at %s", path)
	Koanf = koanf.New(".")

	if err := Koanf.Load(file.Provider(path), json.Parser()); err != nil {
		log.Fatalf("Error ocurred while loading config: %v", err)
	}
	log.Info("Loaded user config file")

}

func LoadDefaultConfig() {
	Koanf = koanf.New(".")

	err := Koanf.Load(rawbytes.Provider(defaultConfig), json.Parser())
	if err != nil {
		log.Fatalf("Error loading default config %v", err)
	}
	log.Info("Loaded default config.")
}
