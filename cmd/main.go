package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/Destinyxus/storeAPI/internal/apiserver"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	server := apiserver.NewAPIServer()
	server.Run()
}
