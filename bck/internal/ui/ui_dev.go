//go:build !prod

package ui

import (
	"net/http"
	"os"
	"path/filepath"
)

func openFileSystem() (http.FileSystem, string, error) {
	candidates := []string{
		filepath.Join("internal", "ui", "dist"),
		filepath.Join("bck", "internal", "ui", "dist"),
		filepath.Join("..", "internal", "ui", "dist"),
	}

	for _, candidate := range candidates {
		absPath, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}

		info, err := os.Stat(absPath)
		if err != nil || !info.IsDir() {
			continue
		}

		if _, err := os.Stat(filepath.Join(absPath, "index.html")); err == nil {
			return http.FS(os.DirFS(absPath)), absPath, nil
		}
	}

	return nil, "", errFrontendBuildNotFound
}
