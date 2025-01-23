package domain

// User represents a user in the system
type User struct {
	ID        string
	Followers []string
	Following []string
}
