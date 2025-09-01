package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/CyberTea0X/codecat/internal/config"
	"github.com/CyberTea0X/codecat/internal/finder"
	"github.com/CyberTea0X/codecat/internal/lang"
)

// Тип для обработки множественных значений флага
type sliceValue []string

func (s *sliceValue) Set(val string) error {
	*s = strings.Split(val, ",")
	return nil
}

func (s *sliceValue) String() string {
	return strings.Join(*s, ",")
}

func main() {
	// --- CLI-флаги ---
	configPath := flag.String("config", "", "path to config.json")

	var langs sliceValue
	flag.Var(&langs, "lang", "comma-separated list of programming languages to search for")

	var ignoreDirs sliceValue
	flag.Var(&ignoreDirs, "I", "comma-separated list of directories to ignore")
	flag.Parse()

	// --- Загрузка конфигурации ---
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// --- Проверка языков ---
	extensions, err := lang.ResolveExtensions(langs, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\nSupported languages: %s\n", err,
			lang.SupportedList(cfg))
		os.Exit(1)
	}

	// --- Поиск и вывод файлов ---
	if err := finder.PrintFiles(".", extensions, ignoreDirs); err != nil {
		fmt.Fprintf(os.Stderr, "Error walking the path: %v\n", err)
		os.Exit(1)
	}
}
