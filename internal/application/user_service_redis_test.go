package application

import (
	"testing"

	"github.com/freischarler/desafio-twitter/internal/domain"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestRedisFollowUser(t *testing.T) {
	mockRedisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	service := NewRedisUserService(mockRedisClient)

	t.Run("should follow user successfully", func(t *testing.T) {
		err := service.FollowUser("1", "2")
		assert.NoError(t, err)
	})

	t.Run("should return error if user tries to follow self", func(t *testing.T) {
		err := service.FollowUser("1", "1")
		assert.Error(t, err)
		assert.Equal(t, domain.ErrCannotFollowSelf, err)
	})
}
