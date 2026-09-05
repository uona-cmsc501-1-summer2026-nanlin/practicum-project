package main

import (
	"log"
	"os"

	"github.com/gobeetle/reply"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project/internal/database"
	"github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project/internal/handlers"
	"github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project/internal/swagger"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "billsplitter.db"
	}

	db, err := database.Connect(dbPath)
	if err != nil {
		log.Fatalf("database: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:      "Split It API",
		ErrorHandler: reply.FiberErrorHandler(),
	})
	app.Use(logger.New())
	app.Use(cors.New())

	swaggerCfg := swagger.DefaultConfig()
	if v := os.Getenv("SWAGGER_BASE_URL"); v != "" {
		swaggerCfg.BaseURL = v
	}
	if v := os.Getenv("OAS_SPEC_FILE"); v != "" {
		swaggerCfg.SpecFile = v
	}
	swagger.MustMount(app, swaggerCfg)

	api := &handlers.API{DB: db}
	api.Register(app)

	app.Static("/app", "./web")
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/app/", fiber.StatusFound)
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return reply.NewFiber(c).JSON(reply.NewData(fiber.Map{"status": "ok"}))
	})

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":55555"
	}
	log.Printf("listening on http://localhost%s (HTTP only, MVP)", addr)
	log.Printf("app UI:     http://localhost%s/app/", addr)
	log.Printf("swagger UI: http://localhost%s/swagger", addr)
	log.Fatal(app.Listen(addr))
}
