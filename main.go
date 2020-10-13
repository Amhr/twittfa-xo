package main

import (
	"OnlineXO/internal/template"
	"OnlineXO/server"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"os"
)

func main() {
	// handling log files
	file, err := os.OpenFile("info.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	e := echo.New()
	e.Renderer = template.NewTemplate()
	//
	//log.SetOutput(file)
	// handling http
	// Middleware
	//e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/xo/ws/:token", server.WebSocket)
	e.GET("/xo", server.Home)
	e.GET("/xo/join/:uid", server.Join)

	e.Static("/xo/static", "static")

	// Start server
	e.Logger.Fatal(e.Start(":8081"))
}
