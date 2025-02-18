package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/freischarler/desafio-twitter/internal/domain"
	"github.com/stretchr/testify/assert"
)

type MockTweetService struct {
	PostTweetFunc   func(userID, tweet string) (string, error)
	GetTimelineFunc func(userID string) ([]domain.Tweet, error)
	GetTweetFunc    func(tweetID string) (domain.Tweet, error)
}

func (m *MockTweetService) PostTweet(userID, tweet string) (string, error) {
	return m.PostTweetFunc(userID, tweet)
}
func (m *MockTweetService) GetTimeline(userID string) ([]domain.Tweet, error) {
	return m.GetTimelineFunc(userID)
}

func (m *MockTweetService) GetTweet(tweetID string) (domain.Tweet, error) {
	return m.GetTweetFunc(tweetID)
}

type MockUserService struct {
	FollowUserFunc func(followerID, followeeID string) error
}

func (m *MockUserService) FollowUser(followerID, followeeID string) error {
	return m.FollowUserFunc(followerID, followeeID)
}

func TestPostTweet(t *testing.T) {
	mockTweetService := &MockTweetService{
		PostTweetFunc: func(userID, tweet string) (string, error) {
			return "12345", nil
		},
	}

	handler := PostTweet(mockTweetService)

	t.Run("should post tweet successfully", func(t *testing.T) {
		reqBody := bytes.NewBufferString(`userID=1&tweet=Hello World`)
		req, err := http.NewRequest("POST", "/tweet", reqBody)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var response map[string]string
		err = json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "Tweet posted successfully", response["message"])
		assert.Equal(t, "12345", response["tweetID"])
	})

	t.Run("should return error if userID is missing", func(t *testing.T) {
		reqBody := bytes.NewBufferString(`tweet=Hello World`)
		req, err := http.NewRequest("POST", "/tweet", reqBody)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "userID is required")
	})

	t.Run("should return error if tweet is missing", func(t *testing.T) {
		reqBody := bytes.NewBufferString(`userID=1`)
		req, err := http.NewRequest("POST", "/tweet", reqBody)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "tweet is required")
	})
}

func TestFollowUser(t *testing.T) {
	mockUserService := &MockUserService{
		FollowUserFunc: func(followerID, followeeID string) error {
			return nil
		},
	}

	handler := FollowUser(mockUserService)

	t.Run("should follow user successfully", func(t *testing.T) {
		reqBody := bytes.NewBufferString(`followerID=1&followeeID=2`)
		req, err := http.NewRequest("POST", "/follow", reqBody)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var response map[string]string
		err = json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, "Followed successfully", response["message"])
	})

	t.Run("should return error if followerID or followeeID is missing", func(t *testing.T) {
		reqBody := bytes.NewBufferString(`followerID=1`)
		req, err := http.NewRequest("POST", "/follow", reqBody)
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Both followerID and followeeID are required")
	})
}

func TestTimeline(t *testing.T) {
	mockTweetService := &MockTweetService{
		GetTimelineFunc: func(userID string) ([]domain.Tweet, error) {
			return []domain.Tweet{
				{TweetID: "1", UserID: "1", Content: "Hello World", Timestamp: time.Now().UnixNano()},
				{TweetID: "2", UserID: "1", Content: "Hello Again", Timestamp: time.Now().UnixNano()},
			}, nil
		},
	}

	handler := Timeline(mockTweetService)

	t.Run("should get timeline successfully", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/timeline/1", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var tweets []domain.Tweet
		err = json.NewDecoder(rr.Body).Decode(&tweets)
		assert.NoError(t, err)
		assert.Len(t, tweets, 2)
		assert.Equal(t, "1", tweets[0].TweetID)
		assert.Equal(t, "2", tweets[1].TweetID)
	})

	t.Run("should return error if userID is missing", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/timeline/", nil)
		assert.NoError(t, err)

		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "userID is required")
	})
}
