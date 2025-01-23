package main

import (
	"log"
	"os"

	"net/http"

	adapterHttp "github.com/freischarler/desafio-twitter/internal/adapters/http"
	"github.com/freischarler/desafio-twitter/internal/infraestructure/redis"
)

func main() {
	redisClient := redis.NewRedisClient()
	http.HandleFunc("/tweet", adapterHttp.PostTweet(redisClient))    // Post a tweet
	http.HandleFunc("/follow", adapterHttp.FollowUser(redisClient))  // Follow a user
	http.HandleFunc("/timeline/", adapterHttp.Timeline(redisClient)) // View timeline

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
