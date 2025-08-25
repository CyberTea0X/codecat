// main.go
// Утилита для вывода содержимого файлов определённых типов.
// Примеры:
//
//	./tool go,js
//	./tool -ext .go,.js,.ts
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
)

const (
	defaultConfigURL = "https://raw.githubusercontent.com/CyberTea0X/codecat/main/config.json"
)

type Config struct {
	Languages map[string][]string `json:"languages"`
	Aliases   map[string]string   `json:"aliases"`
}

var langToExtensions map[string][]string
var aliasMap map[string]string

// Тип для обработки множественных значений флага
type sliceValue []string

func (s *sliceValue) Set(val string) error {
	*s = strings.Split(val, ",")
	return nil
}

func (s *sliceValue) String() string {
	return strings.Join(*s, ",")
}

// configPath возвращает путь к конфигу в домашней директории
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

// downloadDefaultConfig скачивает конфиг с GitHub
func downloadDefaultConfig(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config dir: %v", err)
	}

	resp, err := http.Get(defaultConfigURL)
	if err != nil {
		return fmt.Errorf("failed to download config: %v", err)
	}
	defer resp.Body.Close()

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}
	return nil
}

// loadConfig загружает конфиг из файла, если его нет — скачивает
func loadConfig(configPath string) (*Config, error) {
	var config Config

	if configPath == "" {
		configPath = defaultConfigPath()
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("Config not found, downloading default config to %s\n", configPath)
		if err := downloadDefaultConfig(configPath); err != nil {
			return nil, fmt.Errorf("failed to download default config: %v", err)
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %v", err)
	}

	return &config, nil
}

func extensionsFromLangs(langs []string) []string {
	var supportedExtensions []string
	var unsupportedLangs []string

	for _, lang := range langs {
		lang = strings.ToLower(lang)
		if translatedLang, ok := aliasMap[lang]; ok {
			lang = translatedLang
		}
		exts, ok := langToExtensions[lang]
		if !ok {
			unsupportedLangs = append(unsupportedLangs, lang)
		} else {
			supportedExtensions = append(supportedExtensions, exts...)
		}
	}

	if len(unsupportedLangs) > 0 {
		fmt.Printf("Unsupported languages: %s. Supported languages are: %s\n",
			strings.Join(unsupportedLangs, ", "),
			getSupportedLanguages())
		os.Exit(1)
	}
	return supportedExtensions
}

func main() {
	configPath := flag.String("config", "", "path to config.json")

	var ignoreDirs sliceValue
	flag.Var(&ignoreDirs, "I", "comma-separated list of directories to ignore")

	var extFlag sliceValue
	flag.Var(&extFlag, "ext", "comma-separated list of file extensions to search for (with dot, e.g. .go,.py)")

	flag.Parse()

	// Получаем список языков из аргументов командной строки
	var langs []string
	if flag.NArg() > 0 {
		langs = strings.Split(flag.Arg(0), ",")
	}

	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	langToExtensions = cfg.Languages
	aliasMap = cfg.Aliases

	var supportedExtensions []string
	if len(extFlag) > 0 {
		supportedExtensions = extFlag
	} else if len(langs) > 0 {
		supportedExtensions = extensionsFromLangs(langs)
	} else {
		fmt.Println("Please specify at least one language (as argument) or extension list (-ext).")
		os.Exit(1)
	}

	rootDir := "."
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if slices.Contains(ignoreDirs, info.Name()) {
				fmt.Fprintf(os.Stderr, "Skipping directory: %s\n", path)
				return filepath.SkipDir
			}
			return nil
		}

		fileExt := strings.ToLower(filepath.Ext(path))
		for _, ext := range supportedExtensions {
			if fileExt == ext {
				content, err := os.ReadFile(path)
				if err != nil {
					fmt.Printf("Error reading file %s: %v\n", path, err)
					return nil
				}
				fmt.Printf("// %s\n", path)
				fmt.Println(string(content))
				fmt.Println()
				break
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
		os.Exit(1)
	}
}

func getSupportedLanguages() string {
	processed := make(map[string]bool)
	var list []string
	for lang := range langToExtensions {
		if processed[lang] {
			continue
		}
		var aliases []string
		for alias, mapped := range aliasMap {
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
	return strings.Join(list, ", ")
}
