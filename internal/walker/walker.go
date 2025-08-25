// internal/walker/walker.go
package walker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/CyberTea0X/codecat/internal/config"
	"github.com/CyberTea0X/codecat/internal/progress"
	"github.com/CyberTea0X/codecat/internal/rules"
)

type Config struct {
	RootDir    string // НОВОЕ ПОЛЕ для хранения корневой директории
	ConfigPath string
	SkipExt    string
	MaxSize    string
	Limit      int
	Quiet      bool
	Progress   bool
}

type Walker struct {
	cfg        *Config
	config     *config.Config
	extensions map[string]bool
	rules      *rules.Rules
	progress   *progress.Counter
	limit      int
}

func New(cfg *Config, configData *config.Config, langs, extFlag, ignoreDirs []string, ignoreFiles []string) (*Walker, error) {
	// Убедимся, что RootDir установлен (на случай, если флаг не используется, хотя по умолчанию ".")
	if cfg.RootDir == "" {
		cfg.RootDir = "."
	}
	// Преобразуем в абсолютный путь или нормализуем
	absRootDir, err := filepath.Abs(cfg.RootDir)
	if err != nil {
		return nil, fmt.Errorf("invalid root directory path: %w", err)
	}
	cfg.RootDir = absRootDir

	w := &Walker{
		cfg:        cfg, // Передаем cfg с заполненным RootDir
		config:     configData,
		extensions: make(map[string]bool),
		rules:      rules.New(),
		progress:   progress.New(),
		limit:      cfg.Limit,
	}

	// Игнорируемые файлы
	for _, file := range ignoreFiles {
		w.rules.AddFile(strings.Split(file, ",")...)
	}

	// Игнорируемые директории
	for _, dir := range ignoreDirs {
		w.rules.AddDir(strings.Split(dir, ",")...)
	}

	// Игнорируемые расширения
	if cfg.SkipExt != "" {
		for _, ext := range strings.Split(cfg.SkipExt, ",") {
			w.rules.AddExtension(strings.TrimSpace(ext))
		}
	}

	// Максимальный размер
	if cfg.MaxSize != "" {
		if err := w.rules.SetMaxSize(cfg.MaxSize); err != nil {
			return nil, fmt.Errorf("invalid max-size: %w", err)
		}
	}

	// Определение расширений
	var extensions []string
	var unsupported []string

	if len(extFlag) > 0 {
		extensions = extFlag
	} else if len(langs) > 0 {
		extensions, unsupported = configData.GetExtensions(langs)
		if len(unsupported) > 0 {
			fmt.Fprintf(os.Stderr, "Warning: unsupported languages: %s\n", strings.Join(unsupported, ", "))
		}
	} else {
		return nil, fmt.Errorf("please specify at least one language or extension")
	}

	for _, ext := range extensions {
		w.extensions[strings.ToLower(strings.TrimSpace(ext))] = true
	}

	return w, nil
}

func (w *Walker) Run() error {
	// Используем директорию из конфига
	root := w.cfg.RootDir

	// Проверяем, существует ли директория
	if _, err := os.Stat(root); os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist: %s", root)
	}

	if !w.cfg.Quiet && (w.cfg.Progress || isTerminal()) {
		w.progress.Show()
	}

	var count int
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Может возникнуть ошибка доступа, продолжаем обход
			fmt.Fprintf(os.Stderr, "Warning: skipping %s due to error: %v\n", path, err)
			return nil // filepath.SkipDir или nil, зависит от желаемого поведения при ошибках
		}
		if info.IsDir() {
			// Проверяем, нужно ли пропустить всю директорию
			if w.rules.ShouldSkipDir(path) {
				return filepath.SkipDir
			}
			return nil
		}

		// проверяем лимит перед выводом
		if w.limit > 0 && count >= w.limit {
			return filepath.SkipAll
		}

		if !w.shouldProcessFile(path, info) {
			return nil
		}

		if !w.cfg.Quiet {
			w.progress.Increment()
		}

		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", path, err)
			return nil
		}

		fmt.Printf("// %s\n", path)
		fmt.Println(string(content))
		fmt.Println()

		count++
		return nil
	})

	if !w.cfg.Quiet {
		w.progress.Done()
	}

	return err
}

func (w *Walker) shouldProcessFile(path string, info os.FileInfo) bool {
	// Проверка расширения
	ext := strings.ToLower(filepath.Ext(path))
	if !w.extensions[ext] {
		return false
	}

	// Проверка правил игнорирования
	if w.rules.ShouldSkipFile(path, info) {
		return false
	}

	return true
}

func isTerminal() bool {
	stat, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
