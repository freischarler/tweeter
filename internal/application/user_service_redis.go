package application

import (
	"context"

	"github.com/freischarler/desafio-twitter/internal/domain"
	"github.com/go-redis/redis/v8"
)

// UserService handles user-related operations
type RedisUserService struct {
	RedisClient *redis.Client
	Ctx         context.Context
}

// NewRedisUserService creates a new UserService with a Redis client
func NewRedisUserService(redisClient *redis.Client) *RedisUserService {
	return &RedisUserService{
		RedisClient: redisClient,
		Ctx:         context.Background(),
	}
}

// FollowUser allows a user to follow another user
func (s *RedisUserService) FollowUser(followerID, followeeID string) error {
	if followerID == followeeID {
		return domain.ErrCannotFollowSelf
	}

	err := s.RedisClient.SAdd(s.Ctx, "user:following:"+followerID, followeeID).Err()
	if err != nil {
		return err
	}

	err = s.RedisClient.SAdd(s.Ctx, "user:followers:"+followeeID, followerID).Err()
	if err != nil {
		return err
	}

	return nil
}
