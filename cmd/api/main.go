package main

import (
	"log"

	"go-order-api/internal/app"
)

// @title Order Management API
// @version 1.0
// @description This is a sample Order Management API.
// @host localhost:3333
// @BasePath /
func main() {
	a, err := app.New()
	if err != nil {
		log.Fatalf("Error creating app: %v", err)
	}
	defer a.Close()

	a.Run()
}
