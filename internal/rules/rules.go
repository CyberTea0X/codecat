package rules

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/CyberTea0X/codecat/internal/ignore"
)

type Rules struct {
	skipDirs   map[string]bool
	skipExts   map[string]bool
	skipFiles  map[string]bool
	maxSize    int64
	gitignores []*ignore.Gitignore
}

func New() *Rules {
	return &Rules{
		skipDirs:  make(map[string]bool),
		skipExts:  make(map[string]bool),
		skipFiles: make(map[string]bool),
		maxSize:   -1,
	}
}

func (r *Rules) AddDir(dirs ...string) {
	for _, dir := range dirs {
		r.skipDirs[strings.TrimSpace(dir)] = true
	}
}

func (r *Rules) AddExtension(exts ...string) {
	for _, ext := range exts {
		ext = strings.ToLower(strings.TrimSpace(ext))
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		r.skipExts[ext] = true
	}
}

func (r *Rules) SetMaxSize(sizeStr string) error {
	if sizeStr == "" || sizeStr == "0" {
		r.maxSize = -1
		return nil
	}

	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))
	multiplier := int64(1)

	if strings.HasSuffix(sizeStr, "KB") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
	} else if strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	} else if strings.HasSuffix(sizeStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
	} else if strings.HasSuffix(sizeStr, "B") {
		sizeStr = strings.TrimSuffix(sizeStr, "B")
	}

	size, err := strconv.ParseInt(strings.TrimSpace(sizeStr), 10, 64)
	if err != nil {
		return err
	}

	r.maxSize = size * multiplier
	return nil
}

func (r *Rules) ShouldSkipDir(path string) bool {
	_, dir := filepath.Split(path)
	return r.skipDirs[dir]
}

// Добавляем один хелпер
func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	rel, _ := filepath.Rel(".", p)
	return filepath.ToSlash(rel)
}

func (r *Rules) AddFile(files ...string) {
	for _, file := range files {
		r.skipFiles[normalizePath(file)] = true
	}
}

func (r *Rules) ShouldSkipFile(path string, info os.FileInfo) bool {
	path = normalizePath(path) // <-- добавили
	ext := strings.ToLower(filepath.Ext(path))
	if r.skipExts[ext] {
		return true
	}
	if r.skipFiles[path] {
		return true
	}
	// Проверка размера
	if r.maxSize > 0 && info.Size() > r.maxSize {
		return true
	}

	// Проверка .gitignore
	for _, gi := range r.gitignores {
		if gi.Match(path) {
			return true
		}
	}

	return false
}

func (r *Rules) LoadGitignore(path string) error {
	gi, err := ignore.NewGitignore(path)
	if err != nil {
		return err
	}
	r.gitignores = append(r.gitignores, gi)
	return nil
}
