package ignore

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/CyberTea0X/codecat/internal/fnmatch"
)

type Gitignore struct {
	patterns []pattern
	baseDir  string
}

type pattern struct {
	rule     string
	negated  bool
	dirOnly  bool
	matchDir bool
}

func NewGitignore(path string) (*Gitignore, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	g := &Gitignore{
		baseDir: filepath.Dir(path),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Пропускаем пустые строки и комментарии
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		p := parsePattern(line)
		g.patterns = append(g.patterns, p)
	}

	return g, scanner.Err()
}

func parsePattern(line string) pattern {
	p := pattern{
		rule: line,
	}

	// Убираем пробелы в начале
	line = strings.TrimLeft(line, " \t")

	// Проверка на инверсию
	if strings.HasPrefix(line, "!") {
		p.negated = true
		line = line[1:]
	}

	// Проверка на директории
	if strings.HasSuffix(line, "/") {
		p.dirOnly = true
		line = strings.TrimSuffix(line, "/")
	}

	// Проверка на начало в корне
	if strings.HasPrefix(line, "/") {
		p.matchDir = true
		line = line[1:]
	}

	p.rule = line
	return p
}

func (g *Gitignore) Match(path string) bool {
	// Нормализуем путь
	path = filepath.ToSlash(path)

	// Относительный путь от baseDir
	rel, err := filepath.Rel(g.baseDir, path)
	if err != nil {
		return false
	}
	rel = filepath.ToSlash(rel)

	// Сначала проверяем правила в порядке обратном порядке
	matched := false
	for i := len(g.patterns) - 1; i >= 0; i-- {
		p := g.patterns[i]

		if g.matchPattern(p, rel) {
			if p.negated {
				matched = false
			} else {
				matched = true
			}
		}
	}

	return matched
}

func (g *Gitignore) matchPattern(p pattern, path string) bool {
	// точное совпадение
	if p.rule == path {
		return true
	}
	// fnmatch
	if fnmatch.Match(p.rule, path, fnmatch.FNM_PATHNAME) {
		return true
	}
	// директории
	if p.dirOnly && !strings.HasSuffix(path, "/") {
		return fnmatch.Match(p.rule, path+"/", fnmatch.FNM_PATHNAME)
	}
	return false
}
