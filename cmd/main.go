package main

import (
	"errors"
	"gate-limiter/internal/app"
	"gate-limiter/pkg/redisclient"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func main() {

	// application init
	// environment variable initialization
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}
	// redis initialization
	redisclient.InitRedis()

	// handler
	http.Handle("/", app.InitializeRateHandler())
	err := http.ListenAndServe(":8081", nil)

	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed\n")
	} else if err != nil {
		log.Println("error starting server", err)
		os.Exit(1)
	}
}
