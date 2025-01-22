package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
)

// Contexto global para Redis
var ctx = context.Background()

// Configuración del cliente de Redis con variables de entorno
var redisClient = redis.NewClient(&redis.Options{
	Addr:     os.Getenv("REDIS_HOST"),     // Dirección de Redis desde la variable de entorno
	Password: os.Getenv("REDIS_PASSWORD"), // Contraseña desde la variable de entorno
	DB:       0,                           // Base de datos por defecto
})

const MaxTweetLength = 280

func main() {
	http.HandleFunc("/tweet", PostTweet)    // Publicar un tweet
	http.HandleFunc("/follow", FollowUser)  // Seguir a un usuario
	http.HandleFunc("/timeline/", Timeline) // Ver timeline

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Puerto por defecto
	}

	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}

// PostTweet maneja la publicación de tweets
func PostTweet(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request for %s", r.Method, r.URL.Path)
	r.ParseForm()
	userID := r.FormValue("userID")
	tweet := r.FormValue("tweet")

	// Verificar que se envíen userID y tweet
	if userID == "" {
		http.Error(w, "userID is required", http.StatusBadRequest)
		log.Printf("Failed to post tweet: userID is required")
		return
	}
	if tweet == "" {
		http.Error(w, "tweet is required", http.StatusBadRequest)
		log.Printf("Failed to post tweet: tweet is required")
		return
	}

	if len(tweet) > MaxTweetLength {
		http.Error(w, "Tweet exceeds maximum length", http.StatusBadRequest)
		log.Printf("Failed to post tweet: Tweet exceeds maximum length")
		return
	}

	// Generar un ID único para el tweet
	tweetID := strconv.FormatInt(redisClient.Incr(ctx, "tweetID:counter").Val(), 10)

	// Almacenar el tweet en Redis
	err := redisClient.HSet(ctx, "tweet:"+tweetID, map[string]interface{}{
		"userID":  userID,
		"content": tweet,
	}).Err()
	if err != nil {
		http.Error(w, "Failed to save tweet", http.StatusInternalServerError)
		log.Printf("Failed to save tweet: %v", err)
		return
	}

	// Verificar si el tweet se guardó correctamente
	savedTweet, err := redisClient.HGetAll(ctx, "tweet:"+tweetID).Result()
	if err != nil || len(savedTweet) == 0 {
		http.Error(w, "Failed to verify saved tweet", http.StatusInternalServerError)
		log.Printf("Failed to verify saved tweet: %v", err)
		return
	}
	log.Printf("Tweet saved in Redis: %v", savedTweet)

	// Agregar el tweet al timeline del usuario
	err = redisClient.LPush(ctx, "user:timeline:"+userID, tweetID).Err()
	if err != nil {
		http.Error(w, "Failed to update timeline", http.StatusInternalServerError)
		log.Printf("Failed to update timeline: %v", err)
		return
	}

	// Verificar si el tweet se agregó al timeline del usuario
	timeline, err := redisClient.LRange(ctx, "user:timeline:"+userID, 0, -1).Result()
	if err != nil || len(timeline) == 0 {
		http.Error(w, "Failed to verify timeline update", http.StatusInternalServerError)
		log.Printf("Failed to verify timeline update: %v", err)
		return
	}
	log.Printf("Timeline updated in Redis: %v", timeline)

	response := map[string]string{"message": "Tweet posted successfully", "tweetID": tweetID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("Tweet posted successfully: %s", tweetID)
}

// FollowUser maneja la funcionalidad de seguir a un usuario
func FollowUser(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request for %s", r.Method, r.URL.Path)
	r.ParseForm()
	followerID := r.FormValue("followerID")
	followeeID := r.FormValue("followeeID")

	if followerID == followeeID {
		http.Error(w, "You cannot follow yourself", http.StatusBadRequest)
		log.Printf("Failed to follow user: You cannot follow yourself")
		return
	}

	// Agregar followeeID a la lista de seguidos del followerID
	err := redisClient.SAdd(ctx, "user:following:"+followerID, followeeID).Err()
	if err != nil {
		http.Error(w, "Failed to follow user", http.StatusInternalServerError)
		log.Printf("Failed to follow user: %v", err)
		return
	}

	// Agregar followerID a la lista de seguidores del followeeID
	err = redisClient.SAdd(ctx, "user:followers:"+followeeID, followerID).Err()
	if err != nil {
		http.Error(w, "Failed to follow user", http.StatusInternalServerError)
		log.Printf("Failed to follow user: %v", err)
		return
	}

	response := map[string]string{"message": "Followed successfully"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Printf("User %s followed user %s successfully", followerID, followeeID)
}

// Timeline maneja la visualización del timeline de un usuario
func Timeline(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received %s request for %s", r.Method, r.URL.Path)
	userID := r.URL.Path[len("/timeline/"):]

	// Obtener la lista de usuarios seguidos
	following, err := redisClient.SMembers(ctx, "user:following:"+userID).Result()
	if err != nil {
		http.Error(w, "Failed to fetch following list", http.StatusInternalServerError)
		log.Printf("Failed to fetch following list: %v", err)
		return
	}

	timeline := []map[string]string{}

	// Recopilar tweets de cada usuario seguido
	for _, followeeID := range following {
		tweetIDs, _ := redisClient.LRange(ctx, "user:timeline:"+followeeID, 0, -1).Result()
		for _, tweetID := range tweetIDs {
			tweet, _ := redisClient.HGetAll(ctx, "tweet:"+tweetID).Result()
			if tweet["content"] != "" {
				timeline = append(timeline, map[string]string{
					"userID":  tweet["userID"],
					"content": tweet["content"],
				})
			}
		}
	}

	// Incluir los tweets del propio usuario
	tweetIDs, _ := redisClient.LRange(ctx, "user:timeline:"+userID, 0, -1).Result()
	for _, tweetID := range tweetIDs {
		tweet, _ := redisClient.HGetAll(ctx, "tweet:"+tweetID).Result()
		if tweet["content"] != "" {
			timeline = append(timeline, map[string]string{
				"userID":  tweet["userID"],
				"content": tweet["content"],
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"timeline": timeline})
	log.Printf("Timeline fetched successfully for user %s", userID)
}
