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
	"go":         {".go"},
	"typescript": {".ts", ".tsx"},
	"python":     {".py"},
}

func main() {
	// Флаг для указания языка программирования
	lang := flag.String("lang", "", "programming language to search for (e.g., go, typescript, python)")

	flag.Parse()

	// Проверяем, указан ли язык
	if *lang == "" {
		fmt.Println("Please specify a programming language using the -lang flag.")
		os.Exit(1)
	}

	// Получаем расширения для указанного языка
	extensions, ok := langToExtensions[strings.ToLower(*lang)]
	if !ok {
		fmt.Printf("Unsupported language: %s. Supported languages are: go, typescript, python.\n", *lang)
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
			fileExt := filepath.Ext(path)
			for _, ext := range extensions {
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
	}
}
