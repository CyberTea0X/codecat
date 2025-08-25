// cmd/codecat/main.go
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/CyberTea0X/codecat/internal/config"
	"github.com/CyberTea0X/codecat/internal/walker"
)

const version = "v0.5.1"

func main() {
	// Настройка флагов
	cfg := &walker.Config{} // Этот cfg теперь будет содержать RootDir
	var langsFlag, extFlag multiFlag
	var ignoreDirs multiFlag
	var ignoreFiles multiFlag

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `Codecat v%s - Concatenate source files

USAGE:
    codecat [OPTIONS] [LANGUAGES...]

EXAMPLES:
    codecat go,js
    codecat -l go -l js
    codecat --ext .go,.js --limit 10
    codecat --skip-ext .o,.a --max-size 1MB
    codecat -d /path/to/project go

OPTIONS:
`, version)
		flag.PrintDefaults()
	}

	// Флаги
	flag.Var(&langsFlag, "l", "Language to include (can be used multiple times)")
	flag.Var(&langsFlag, "lang", "Language to include (can be used multiple times)")
	flag.Var(&extFlag, "ext", "File extensions to include (comma-separated)")
	flag.Var(&ignoreDirs, "I", "Directories to ignore (comma-separated)")
	flag.Var(&ignoreFiles, "i", "Files to ignore (comma-separated)")

	// НОВЫЙ ФЛАГ -d
	flag.StringVar(&cfg.RootDir, "d", ".", "Directory to scan") // Устанавливаем "." по умолчанию

	flag.StringVar(&cfg.ConfigPath, "config", "", "Path to config.json")
	flag.StringVar(&cfg.SkipExt, "skip-ext", "", "Extensions to skip (comma-separated)")
	flag.StringVar(&cfg.MaxSize, "max-size", "0", "Maximum file size (e.g., 1MB, 500KB)")
	flag.IntVar(&cfg.Limit, "limit", 0, "Maximum number of files to output")
	flag.BoolVar(&cfg.Quiet, "q", false, "Quiet mode (no progress)")
	flag.BoolVar(&cfg.Progress, "p", false, "Show progress")

	var showVersion bool
	flag.BoolVar(&showVersion, "v", false, "Show version")
	flag.BoolVar(&showVersion, "version", false, "Show version")

	flag.Parse()

	if showVersion {
		fmt.Printf("codecat version %s\n", version)
		os.Exit(0)
	}

	// Обработка языков из позиционных аргументов
	langs := langsFlag
	for _, arg := range flag.Args() {
		langs = append(langs, strings.Split(arg, ",")...)
	}

	// Загрузка конфигурации
	cfgData, err := config.Load(cfg.ConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Создание walker
	w, err := walker.New(cfg, cfgData, langs, extFlag, ignoreDirs, ignoreFiles)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Запуск
	if err := w.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// multiFlag позволяет использовать флаг многократно
type multiFlag []string

func (m *multiFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *multiFlag) Set(value string) error {
	if value == "" {
		return nil
	}
	*m = append(*m, strings.Split(value, ",")...)
	return nil
}
