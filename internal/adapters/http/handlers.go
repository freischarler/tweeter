package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/freischarler/desafio-twitter/internal/application"
	"github.com/freischarler/desafio-twitter/internal/domain"
	"github.com/go-redis/redis/v8"
)

// PostTweet handles posting a tweet
func PostTweet(redisClient *redis.Client) http.HandlerFunc {
	tweetService := application.NewTweetService(redisClient)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request for %s", r.Method, r.URL.Path)
		r.ParseForm()
		userID := r.FormValue("userID")
		tweet := r.FormValue("tweet")

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

		tweetID, err := tweetService.PostTweet(userID, tweet)
		if err != nil {
			http.Error(w, "Failed to save tweet", http.StatusInternalServerError)
			log.Printf("Failed to save tweet: %v", err)
			return
		}

		response := map[string]string{"message": "Tweet posted successfully", "tweetID": tweetID}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		log.Printf("Tweet posted successfully: %s", tweetID)
	}
}

// FollowUser handles following a user
func FollowUser(redisClient *redis.Client) http.HandlerFunc {
	userService := application.NewUserService(redisClient)

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request for %s", r.Method, r.URL.Path)
		r.ParseForm()
		followerID := r.FormValue("followerID")
		followeeID := r.FormValue("followeeID")

		if followerID == "" || followeeID == "" {
			http.Error(w, "Both followerID and followeeID are required", http.StatusBadRequest)
			log.Printf("Failed to follow user: Missing followerID or followeeID")
			return
		}

		if err := userService.FollowUser(followerID, followeeID); err != nil {
			http.Error(w, "Failed to follow user", http.StatusInternalServerError)
			log.Printf("Failed to follow user: %v", err)
			return
		}

		response := map[string]string{"message": "Followed successfully"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		log.Printf("User %s followed user %s successfully", followerID, followeeID)
	}
}

// Timeline handles viewing a user's timeline
func Timeline(redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request for %s", r.Method, r.URL.Path)
		userID := r.URL.Path[len("/timeline/"):]

		if userID == "" {
			http.Error(w, "userID is required", http.StatusBadRequest)
			log.Printf("Failed to fetch timeline: userID is required")
			return
		}

		// Fetch the list of followed users
		following, err := redisClient.SMembers(context.Background(), "user:following:"+userID).Result()
		if err != nil {
			http.Error(w, "Failed to fetch following list", http.StatusInternalServerError)
			log.Printf("Failed to fetch following list: %v", err)
			return
		}

		var timeline []domain.Tweet

		// Collect tweets from each followed user
		for _, followeeID := range following {
			tweetIDs, _ := redisClient.LRange(context.Background(), "user:timeline:"+followeeID, 0, -1).Result()
			for _, tweetID := range tweetIDs {
				tweetData, _ := redisClient.HGetAll(context.Background(), "tweet:"+tweetID).Result()
				if tweetData["content"] != "" {
					timestamp, _ := strconv.ParseInt(tweetData["timestamp"], 10, 64)
					timeline = append(timeline, domain.Tweet{
						UserID:    tweetData["userID"],
						Content:   tweetData["content"],
						Timestamp: timestamp,
					})
				}
			}
		}

		// Include the user's own tweets
		tweetIDs, _ := redisClient.LRange(context.Background(), "user:timeline:"+userID, 0, -1).Result()
		for _, tweetID := range tweetIDs {
			tweetData, _ := redisClient.HGetAll(context.Background(), "tweet:"+tweetID).Result()
			if tweetData["content"] != "" {
				timestamp, _ := strconv.ParseInt(tweetData["timestamp"], 10, 64)
				timeline = append(timeline, domain.Tweet{
					UserID:    tweetData["userID"],
					Content:   tweetData["content"],
					Timestamp: timestamp,
				})
			}
		}

		// Sort tweets by timestamp
		sort.Slice(timeline, func(i, j int) bool {
			return timeline[i].Timestamp > timeline[j].Timestamp
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"timeline": timeline})
		log.Printf("Timeline fetched successfully for user %s", userID)
	}
}
