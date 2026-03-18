//go:build prod

package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var embeddedDist embed.FS

func openFileSystem() (http.FileSystem, string, error) {
	distFS, err := fs.Sub(embeddedDist, "dist")
	if err != nil {
		return nil, "", err
	}

	return http.FS(distFS), "embedded frontend assets", nil
}
