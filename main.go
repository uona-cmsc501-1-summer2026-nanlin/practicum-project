package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project/internal/database"
	"github.com/uona-cmsc501-1-summer2026-nanlin/practicum-project/internal/handlers"
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
		AppName: "Shared Bill Splitter API",
	})
	app.Use(logger.New())
	app.Use(cors.New())

	api := &handlers.API{DB: db}
	api.Register(app)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":3000"
	}
	log.Printf("listening on http://localhost%s (HTTP only, MVP)", addr)
	log.Fatal(app.Listen(addr))
}
