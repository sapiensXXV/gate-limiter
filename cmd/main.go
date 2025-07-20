package main

import (
	"errors"
	"fmt"
	"gate-limiter/internal/limiter"
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
	http.HandleFunc("/", limiter.HandleRateLimit)
	err := http.ListenAndServe(":8081", nil)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Println("server closed\n")
	} else if err != nil {
		fmt.Println("error starting server", err)
		os.Exit(1)
	}
}
