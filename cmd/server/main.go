package main

import (
	"log"
	"net/http"
	"os"
	"time"

	adapterHttp "github.com/freischarler/desafio-twitter/internal/adapters/http"
	"github.com/freischarler/desafio-twitter/internal/application"
	"github.com/freischarler/desafio-twitter/internal/infraestructure/redis"
	"github.com/freischarler/desafio-twitter/internal/middleware"
)

func main() {
	var tweetService application.TweetService
	var userService application.UserService

	redisClient := redis.NewRedisClient()
	tweetService = application.NewRedisTweetService(redisClient)
	userService = application.NewRedisUserService(redisClient)

	mux := http.NewServeMux()

	mux.HandleFunc("/tweet", adapterHttp.PostTweet(tweetService))    // Post a tweet
	mux.HandleFunc("/follow", adapterHttp.FollowUser(userService))   // Follow a user
	mux.HandleFunc("/timeline/", adapterHttp.Timeline(tweetService)) // View timeline

	// Apply rate limiting middleware
	rateLimitMiddleware := middleware.RateLimitMiddleware(time.Minute, 100)
	handler := rateLimitMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}

	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
