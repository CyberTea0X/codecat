package lang

import (
	"fmt"
	"strings"

	"github.com/CyberTea0X/codecat/internal/config"
)

// ResolveExtensions превращает список языков в список расширений.
func ResolveExtensions(langs []string, cfg *config.Config) ([]string, error) {
	var extensions []string
	var unsupported []string

	for _, l := range langs {
		l = strings.ToLower(l)
		if alias, ok := cfg.Aliases[l]; ok {
			l = alias
		}
		exts, ok := cfg.Languages[l]
		if !ok {
			unsupported = append(unsupported, l)
			continue
		}
		extensions = append(extensions, exts...)
	}

	if len(unsupported) > 0 {
		return nil, fmt.Errorf("unsupported languages: %s", strings.Join(unsupported, ", "))
	}
	return extensions, nil
}

// SupportedList возвращает строку вида "go (golang), typescript (ts, tsx), ...".
func SupportedList(cfg *config.Config) string {
	seen := make(map[string]bool)
	var parts []string

	for lang := range cfg.Languages {
		if seen[lang] {
			continue
		}
		seen[lang] = true

		var aliases []string
		for alias, mapped := range cfg.Aliases {
			if mapped == lang && alias != lang {
				aliases = append(aliases, alias)
			}
		}

		if len(aliases) > 0 {
			parts = append(parts, fmt.Sprintf("%s (%s)", lang, strings.Join(aliases, ", ")))
		} else {
			parts = append(parts, lang)
		}
	}
	return strings.Join(parts, ", ")
}
