package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/freischarler/desafio-twitter/internal/domain"
)

// PostTweet handles posting a tweet
func PostTweet(tweetService domain.TweetService) http.HandlerFunc {
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
func FollowUser(userService domain.UserService) http.HandlerFunc {
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
func Timeline(tweetService domain.TweetService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received %s request for %s", r.Method, r.URL.Path)
		userID := r.URL.Path[len("/timeline/"):]

		if userID == "" {
			http.Error(w, "userID is required", http.StatusBadRequest)
			log.Printf("Failed to fetch timeline: userID is required")
			return
		}

		tweets, err := tweetService.GetTimeline(userID)
		if err != nil {
			http.Error(w, "Failed to fetch timeline", http.StatusInternalServerError)
			log.Printf("Failed to fetch timeline: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tweets)
		log.Printf("Fetched timeline for user %s successfully", userID)
	}
}
