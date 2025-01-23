package application

// UserService defines the interface for user-related operations
type UserService interface {
	FollowUser(followerID, followeeID string) error
}
