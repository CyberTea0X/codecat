package finder

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// PrintFiles рекурсивно выводит содержимое файлов с нужными расширениями.
func PrintFiles(root string, exts []string, ignoreDirs []string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if slices.Contains(ignoreDirs, d.Name()) {
				fmt.Fprintf(os.Stderr, "Skipping directory: %s\n", path)
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		for _, e := range exts {
			if ext == e {
				data, err := os.ReadFile(path)
				if err != nil {
					fmt.Printf("Error reading file %s: %v\n", path, err)
					return nil
				}
				fmt.Printf("// %s\n%s\n\n", path, string(data))
				break
			}
		}
		return nil
	})
}
