package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const defaultConfigURL = "https://raw.githubusercontent.com/CyberTea0X/codecat/main/config.json"

type Config struct {
	Languages map[string][]string `json:"languages"`
	Aliases   map[string]string   `json:"aliases"`
}

// Load загружает конфигурацию
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = defaultConfigPath()
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("Config not found, downloading default config to %s\n", configPath)
		if err := downloadDefaultConfig(configPath); err != nil {
			return nil, fmt.Errorf("failed to download default config: %w", err)
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// GetExtensions возвращает расширения для языков
func (c *Config) GetExtensions(langs []string) ([]string, []string) {
	var extensions []string
	var unsupported []string
	processed := make(map[string]bool)

	for _, lang := range langs {
		lang = strings.ToLower(strings.TrimSpace(lang))
		if translated, ok := c.Aliases[lang]; ok {
			lang = translated
		}

		if processed[lang] {
			continue
		}

		if exts, ok := c.Languages[lang]; ok {
			extensions = append(extensions, exts...)
			processed[lang] = true
		} else {
			unsupported = append(unsupported, lang)
		}
	}

	return extensions, unsupported
}

// GetSupportedLanguages возвращает список поддерживаемых языков
func (c *Config) GetSupportedLanguages() string {
	processed := make(map[string]bool)
	var list []string

	for lang := range c.Languages {
		if processed[lang] {
			continue
		}
		var aliases []string
		for alias, mapped := range c.Aliases {
			if mapped == lang && alias != lang {
				aliases = append(aliases, alias)
			}
		}
		if len(aliases) > 0 {
			list = append(list, fmt.Sprintf("%s (%s)", lang, strings.Join(aliases, ", ")))
		} else {
			list = append(list, lang)
		}
		processed[lang] = true
	}
	sort.Strings(list)
	return strings.Join(list, ", ")
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic("can't get user home directory: " + err.Error())
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(home, "Documents", "codecat", "config.json")
	}
	return filepath.Join(home, ".config", "codecat", "config.json")
}

func downloadDefaultConfig(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	resp, err := http.Get(defaultConfigURL)
	if err != nil {
		return fmt.Errorf("failed to download config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status downloading config: %s", resp.Status)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}
