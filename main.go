package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	_ "embed"
)

//go:embed config.json
var defaultConfig []byte

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

func loadConfig(configPath string) (*Config, error) {
	var config Config

	if configPath == "" {
		// Используем embed-конфиг
		err := json.Unmarshal(defaultConfig, &config)
		if err != nil {
			return nil, fmt.Errorf("failed to parse embedded config: %v", err)
		}
	} else {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %v", err)
		}
		err = json.Unmarshal(data, &config)
		if err != nil {
			return nil, fmt.Errorf("failed to parse config: %v", err)
		}
	}

	return &config, nil
}

func main() {
	// Флаг для указания пути к конфигурационному файлу
	configPath := flag.String("config", "", "path to config.json")

	// Флаг для указания языков программирования
	var langs sliceValue
	flag.Var(&langs, "lang", "comma-separated list of programming languages to search for (supports aliases like ts, js, py, etc.)")

	// Флаг для игнорирования директорий
	var ignoreDirs sliceValue
	flag.Var(&ignoreDirs, "I", "comma-separated list of directories to ignore")

	flag.Parse()

	// Загружаем конфигурацию
	cfg, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	langToExtensions = cfg.Languages
	aliasMap = cfg.Aliases

	// Проверяем, указан ли хотя бы один язык
	if len(langs) == 0 {
		fmt.Println("Please specify at least one programming language using the -lang flag.")
		os.Exit(1)
	}

	// Проверяем поддержку указанных языков и собираем расширения
	var supportedExtensions []string
	var unsupportedLangs []string

	for _, lang := range langs {
		lang = strings.ToLower(lang)

		// Проверяем алиасы
		if translatedLang, ok := aliasMap[lang]; ok {
			lang = translatedLang
		}

		extensions, ok := langToExtensions[lang]
		if !ok {
			unsupportedLangs = append(unsupportedLangs, lang)
		} else {
			supportedExtensions = append(supportedExtensions, extensions...)
		}
	}

	// Если есть неподдерживаемые языки, выводим ошибку
	if len(unsupportedLangs) > 0 {
		fmt.Printf("Unsupported languages: %s. Supported languages are: %s\n",
			strings.Join(unsupportedLangs, ", "),
			getSupportedLanguages())
		os.Exit(1)
	}

	// Текущая директория, из которой запущена программа
	rootDir := "."

	// Рекурсивный обход директорий
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Пропускаем игнорируемые директории
		if info.IsDir() {
			if slices.Contains(ignoreDirs, info.Name()) {
				fmt.Fprintf(os.Stderr, "Skipping directory: %s\n", path)
				return filepath.SkipDir
			}
		}

		// Проверяем, является ли это файлом с нужным расширением
		if !info.IsDir() {
			fileExt := strings.ToLower(filepath.Ext(path))
			for _, ext := range supportedExtensions {
				if fileExt == ext {
					// Читаем содержимое файла
					content, err := os.ReadFile(path)
					if err != nil {
						fmt.Printf("Error reading file %s: %v\n", path, err)
						return nil
					}

					// Выводим комментарий с именем файла
					fmt.Printf("// %s\n", path)

					// Выводим содержимое файла
					fmt.Println(string(content))
					fmt.Println() // Добавляем пустую строку для разделения файлов
					break
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
		os.Exit(1)
	}
}

// Вспомогательная функция для получения списка поддерживаемых языков с алиасами
func getSupportedLanguages() string {
	// Используем map для отслеживания уже добавленных языков
	processedLangs := make(map[string]bool)
	var result []string

	for lang := range langToExtensions {
		// Пропускаем дубликаты
		if processedLangs[lang] {
			continue
		}

		// Собираем все алиасы для текущего языка
		var aliases []string
		for alias, mappedLang := range aliasMap {
			if mappedLang == lang && alias != lang {
				aliases = append(aliases, alias)
			}
		}

		// Добавляем язык и его алиасы
		if len(aliases) > 0 {
			result = append(result, fmt.Sprintf("%s (%s)", lang, strings.Join(aliases, ", ")))
		} else {
			result = append(result, lang)
		}

		processedLangs[lang] = true
	}

	return strings.Join(result, ", ")
}
