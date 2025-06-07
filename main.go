package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Маппинг языков к их расширениям
var langToExtensions = map[string][]string{
	"golang":      {".go"},
	"typescript":  {".ts", ".tsx"},
	"javascript":  {".js", ".jsx"},
	"python":      {".py"},
	"java":        {".java"},
	"c++":         {".cpp", ".hpp", ".h"},
	"ruby":        {".rb"},
	"php":         {".php"},
	"swift":       {".swift"},
	"kotlin":      {".kt", ".kts"},
	"rust":        {".rs"},
	"sql":         {".sql"},
	"c#":          {".cs"},
	"c":           {".c", ".h"},
	"bash":        {".sh"},
	"powershell":  {".ps1"},
	"scala":       {".scala"},
	"r":           {".r"},
	"perl":        {".pl"},
	"dart":        {".dart"},
	"lua":         {".lua"},
	"groovy":      {".groovy"},
	"assembly":    {".asm", ".s"},
	"objective-c": {".m", ".mm"},
	"json":        {".json"},
	"yaml":        {".yaml", ".yml"},
	"toml":        {".toml"},
	"html":        {".html"},
	"css":         {".css"},
}

// Маппинг алиасов к основным названиям языков
var aliasMap = map[string]string{
	"ts":     "typescript",
	"js":     "javascript",
	"py":     "python",
	"go":     "golang",
	"java":   "java",
	"cpp":    "c++",
	"c++":    "c++",
	"cc":     "c++",
	"rb":     "ruby",
	"php":    "php",
	"swift":  "swift",
	"kt":     "kotlin",
	"kts":    "kotlin",
	"rs":     "rust",
	"sql":    "sql",
	"cs":     "c#",
	"c#":     "c#",
	"c":      "c",
	"h":      "c",
	"sh":     "bash",
	"ps1":    "powershell",
	"scala":  "scala",
	"r":      "r",
	"pl":     "perl",
	"dart":   "dart",
	"lua":    "lua",
	"groovy": "groovy",
	"asm":    "assembly",
	"s":      "assembly",
	"objc":   "objective-c",
	"mm":     "objective-c",
	"m":      "objective-c",
}

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
	// Флаг для указания языков программирования
	var langs sliceValue
	flag.Var(&langs, "lang", "comma-separated list of programming languages to search for (supports aliases like ts, js, py, etc.)")

	flag.Parse()

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
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
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
