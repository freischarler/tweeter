package domain

import "errors"

// Define domain-specific errors
var (
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
	ErrTweetTooLong     = errors.New("tweet is too long")
)
