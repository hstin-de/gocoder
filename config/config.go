package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

var (
	Languages              []string = []string{"en", "de", "fr", "es", "it", "nl", "pt", "ru", "zh"}
	WikimediaMaxImportance float64  = 500.0
	Planet                 string   = ""
	WhosOnFirst            string   = ""
	WikimediaImportance    string   = ""

	// INTERMEDIATES
	BoundingBoxes string = ""
	Countries     string = ""

	Output        string = "geocoder.gpkg"
	Database      string = "geocoder.gpkg"
	EnableForward bool   = true
	EnableReverse bool   = true
)

type jsonConfig struct {
	Languages              []string `json:"languages,omitempty"`
	WikimediaMaxImportance *float64 `json:"wikimedia_max_importance,omitempty"`
	Planet                 string   `json:"planet,omitempty"`
	WhosOnFirst            string   `json:"whos_on_first,omitempty"`
	WikimediaImportance    string   `json:"wikimedia_importance,omitempty"`
	Output                 string   `json:"output,omitempty"`
	Database               string   `json:"database,omitempty"`
	EnableForward          *bool    `json:"enable_forward,omitempty"`
	EnableReverse          *bool    `json:"enable_reverse,omitempty"`
}

func init() {
	Load()
}

func Load() {
	godotenv.Load()

	if data, err := os.ReadFile("config.json"); err == nil {
		var cfg jsonConfig
		if json.Unmarshal(data, &cfg) == nil {
			if cfg.Languages != nil {
				Languages = cfg.Languages
			}
			if cfg.WikimediaMaxImportance != nil {
				WikimediaMaxImportance = *cfg.WikimediaMaxImportance
			}
			if cfg.Planet != "" {
				Planet = cfg.Planet
			}
			if cfg.WhosOnFirst != "" {
				WhosOnFirst = cfg.WhosOnFirst
			}
			if cfg.WikimediaImportance != "" {
				WikimediaImportance = cfg.WikimediaImportance
			}
			if cfg.Output != "" {
				Output = cfg.Output
			}
			if cfg.Database != "" {
				Database = cfg.Database
			}
			if cfg.EnableForward != nil {
				EnableForward = *cfg.EnableForward
			}
			if cfg.EnableReverse != nil {
				EnableReverse = *cfg.EnableReverse
			}
		}
	}

	if val := os.Getenv("LANGUAGES"); val != "" {
		Languages = strings.Split(val, ",")
	}
	if val := os.Getenv("WIKIMEDIA_MAX_IMPORTANCE"); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			WikimediaMaxImportance = f
		}
	}
	if val := os.Getenv("PLANET"); val != "" {
		Planet = val
	}
	if val := os.Getenv("WHOS_ON_FIRST"); val != "" {
		WhosOnFirst = val
	}
	if val := os.Getenv("WIKIMEDIA_IMPORTANCE"); val != "" {
		WikimediaImportance = val
	}
	if val := os.Getenv("OUTPUT"); val != "" {
		Output = val
	}
	if val := os.Getenv("DATABASE"); val != "" {
		Database = val
	}
	if val := os.Getenv("ENABLE_FORWARD"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			EnableForward = b
		}
	}
	if val := os.Getenv("ENABLE_REVERSE"); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			EnableReverse = b
		}
	}

	// INTERMEDIATES
	OutputPath := filepath.Dir(Output)

	BoundingBoxes = filepath.Join(OutputPath, "bounding_boxes.geo")
	Countries = filepath.Join(OutputPath, "countries.geojson")
}
