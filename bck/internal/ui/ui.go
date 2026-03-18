package ui

import (
	"errors"
	"log"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	fiberfilesystem "github.com/gofiber/fiber/v2/middleware/filesystem"
)

var errFrontendBuildNotFound = errors.New("frontend build not found")

// Mount exposes the frontend app while keeping API routes untouched.
func Mount(app *fiber.App, apiPrefix string) {
	root, source, err := openFileSystem()
	if err != nil {
		log.Printf("frontend: %v. Run `npm run build` from the repository root.", err)
		mountUnavailable(app, apiPrefix)
		return
	}

	app.Use("/", fiberfilesystem.New(fiberfilesystem.Config{
		Root:  root,
		Index: "index.html",
		Next: func(c *fiber.Ctx) bool {
			return isAPIPath(c.Path(), apiPrefix)
		},
	}))

	app.Get("*", func(c *fiber.Ctx) error {
		path := c.Path()
		if isAPIPath(path, apiPrefix) {
			return fiber.ErrNotFound
		}
		if filepath.Ext(path) != "" {
			return fiber.ErrNotFound
		}
		return fiberfilesystem.SendFile(c, root, "/index.html")
	})

	log.Printf("frontend: serving static files from %s", source)
}

func mountUnavailable(app *fiber.App, apiPrefix string) {
	app.Get("*", func(c *fiber.Ctx) error {
		path := c.Path()
		if isAPIPath(path, apiPrefix) {
			return fiber.ErrNotFound
		}
		if filepath.Ext(path) != "" {
			return fiber.ErrNotFound
		}

		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"message": "frontend build not found",
			"hint":    "run `npm run build` from the repository root to generate bck/internal/ui/dist",
		})
	})
}

func isAPIPath(path, apiPrefix string) bool {
	apiPrefix = strings.TrimSpace(apiPrefix)
	if apiPrefix == "" || apiPrefix == "/" {
		return false
	}

	return path == apiPrefix || strings.HasPrefix(path, apiPrefix+"/")
}
