package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Languages    map[string][]string `json:"languages"`
	Aliases      map[string]string   `json:"aliases"`
	IgnoreHidden bool                `json:"ignore_hidden"`
	IgnoreDirs   []string            `json:"ignore_dirs"`
}

// Load загружает конфиг:
// 1. Если передан явный путь — используем его.
// 2. Иначе ищем в системной директории, при необходимости скачивая.
// 3. Если и там нет — используем встроенный.
func Load(explicitPath string) (*Config, error) {
	var data []byte
	var err error

	switch {
	case explicitPath != "":
		data, err = os.ReadFile(explicitPath)
	case true:
		sysPath, err2 := ensureConfigDir()
		if err2 != nil {
			return nil, err2
		}
		if errDl := downloadConfigIfNotExists(sysPath); errDl != nil {
			return nil, err
		} else {
			data, err = os.ReadFile(sysPath)
		}
	}

	if err != nil && len(data) == 0 {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}
