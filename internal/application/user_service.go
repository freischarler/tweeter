package application

import (
	"context"

	"github.com/freischarler/desafio-twitter/internal/domain"
	"github.com/go-redis/redis/v8"
)

// UserService handles user-related operations
type UserService struct {
	RedisClient *redis.Client
	Ctx         context.Context
}

// NewUserService creates a new UserService
func NewUserService(redisClient *redis.Client) *UserService {
	return &UserService{
		RedisClient: redisClient,
		Ctx:         context.Background(),
	}
}

// FollowUser allows a user to follow another user
func (s *UserService) FollowUser(followerID, followeeID string) error {
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
