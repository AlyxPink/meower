package main

import (
	"log"

	"github.com/AlyxPink/meower/internal/pkg/web/handlers"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	app := handlers.NewServer()
	app.Listen("0.0.0.0:3000")
}
