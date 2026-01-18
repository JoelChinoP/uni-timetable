package pkg

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// SetupLogging configura el logger de Fiber usando exclusivamente logger.New(...)
// - Columnas “alineadas” (hasta donde permite el Format string)
// - Latencia en milisegundos (custom tag)
// - Timestamp con milisegundos
func SetupLogging(app *fiber.App, environment string) {
	if environment != "development" {
		// En prod puedes mandarlo a archivo o a un writer; por ahora, no log.
		app.Use(logger.New(logger.Config{Output: io.Discard}))
		return
	}

	logConfig := logger.Config{
		Format:     "${time} | ${status}\t | ${ip}\t | ${method}\t | ${path} | ${latency_ms}\n",
		TimeFormat: "2006-01-02 15:04:05.000", // milisegundos
		TimeZone:   "America/Lima",
		Output:     os.Stdout,
		CustomTags: map[string]logger.LogFunc{
			"latency_ms": func(out logger.Buffer, c *fiber.Ctx, data *logger.Data, extra string) (int, error) {
				// fasthttp guarda el inicio del request aquí
				start := c.Context().Time()
				d := time.Since(start)
				ms := float64(d) / float64(time.Millisecond)
				return out.WriteString(fmt.Sprintf("%.3fms", ms))
			},
		},
	}

	// Debe registrarse de los primeros middlewares
	app.Use(logger.New(logConfig))
}
