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
	ignoreHidden := flag.Bool("hidden", false, "include hidden files/dirs")
	rootDir := flag.String("d", ".", "root directory to scan (default \".\")")

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

	if ignoreDirs == nil {
		ignoreDirs = cfg.IgnoreDirs
	}

	if ignoreHidden == nil {
		ignoreHidden = &cfg.IgnoreHidden
	}

	// --- Поиск и вывод файлов ---
	if err := finder.PrintFiles(*rootDir, extensions, ignoreDirs, *ignoreHidden); err != nil {
		fmt.Fprintf(os.Stderr, "Error walking the path: %v\n", err)
		os.Exit(1)
	}
}
