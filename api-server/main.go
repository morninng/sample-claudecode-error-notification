package main

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/hello", handleHello)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting api-server on port %s", port)
	e.Logger.Fatal(e.Start(":" + port))
}

func handleHello(c echo.Context) error {
	message := c.QueryParam("message")

	// Check if message parameter is "error"
	if message == "error" {
		c.Logger().Error("error occurred with invalid query message")
		return c.String(http.StatusBadRequest, "error occured with invalid query message")
	}

	// Default response
	return c.String(http.StatusOK, "hello world")
}
