package utils

import (
	"bytes"
	_ "embed"
	"os"

	"github.com/charmbracelet/log"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

var (
	// Config manager
	Koanf *koanf.Koanf
)

//go:embed config/default.json
var defaultConfig []byte

// Takes the bytes from a JSON array and removes their comment lines (lines starting with //)
func StripCommentsFromJSON(fileContent []byte) ([]byte, error) {
	lines := bytes.Split(fileContent, []byte("\n"))
	var filteredLines [][]byte

	for _, line := range lines {
		trimmedLine := bytes.TrimSpace(line)
		if !bytes.HasPrefix(trimmedLine, []byte("//")) {
			filteredLines = append(filteredLines, line)
		}
	}

	return bytes.Join(filteredLines, []byte("\n")), nil
}

// Loads the user defined config.
func LoadUserConfig(path string) {
	log.Debug("Loading configuration at %s", path)
	if Koanf == nil {
		Koanf = koanf.New(".")
	}

	fileContent, err := os.ReadFile(path)
	if err != nil {
		log.Warnf("Unable to read user config file, %v", err)
		return
	}

	sanitizedJSON, err := StripCommentsFromJSON(fileContent)
	if err != nil {
		log.Warnf("Unable to strip comments from user config file, %v", err)
		return
	}

	if err := Koanf.Load(rawbytes.Provider(sanitizedJSON), json.Parser()); err != nil {
		log.Fatalf("Error occurred while loading config: %v", err)
	}
	log.Info("Loaded user config file")
}

// Loads the default user config
func LoadDefaultConfig() {
	if Koanf == nil {
		Koanf = koanf.New(".")
	}

	sanitizedJSON, err := StripCommentsFromJSON(defaultConfig)
	if err != nil {
		log.Warnf("Unable to read user config file, %v", err)
		return
	}

	err = Koanf.Load(rawbytes.Provider(sanitizedJSON), json.Parser())
	if err != nil {
		log.Fatalf("Error loading default config %v", err)
	}
	log.Info("Loaded default config.")
}
