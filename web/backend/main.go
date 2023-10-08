package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	app := fiber.New()

	// Logging for each request
	app.Use(logger.New())

	// Serve the Svelte built app
	app.Static("/", "../frontend/public")

	// Start the Fiber server
	log.Fatal(app.Listen(":3000"))
}
