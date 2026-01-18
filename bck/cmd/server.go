package main

import (
	"log"

	"github.com/JoelChinoP/timetable_bck/internal/http"
	/* 	"github.com/JoelChinoP/timetable_bck/internal/database"*/
	"github.com/JoelChinoP/timetable_bck/pkg"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Cargar configuración de la aplicación
	cfg := pkg.MustLoadConfig()

	// Inicializar la base de datos
	/* ctx := context.Background()
	database.InitDB(ctx, cfg.Database)
	defer database.Close() */

	// Crear instancia de Fiber con configuración personalizada
	app := fiber.New(fiber.Config{
		CaseSensitive:                true, // Habilita la sensibilidad a mayúsculas y minúsculas en las rutas
		StrictRouting:                true, // Habilita el enrutamiento estricto
		DisablePreParseMultipartForm: true, // Desactiva el preparseo de formularios multipart
		EnablePrintRoutes:            true, // Imprime las rutas al iniciar el servidor
		AppName:                      cfg.AppName,
		ErrorHandler:                 pkg.SetupGlobalErrorHandler, // Configura el manejador de errores global
	})

	// Configurar middlewares
	pkg.SetupCORS(app, cfg.CORSOrigins) // Configura CORS para la aplicación
	pkg.SetupLogging(app, cfg.Env)      // Configura el middleware de registro para registrar las solicitudes y respuestas

	// Configurar rutas
	http.SetupRoutes(app) // Configura las rutas de la API

	// Iniciar servidor
	addr := ":" + cfg.Port
	log.Printf("%s (%s) listening on %s", cfg.AppName, cfg.Env, addr)
	log.Fatal(app.Listen(addr))
}
