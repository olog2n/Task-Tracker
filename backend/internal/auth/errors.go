package auth

import "errors"

var (
	ErrTokenBlacklist          = errors.New("Token blacklist error")
	ErrTokenHasBeenRevoked     = errors.New("Token has been revoked")
	ErrInvalidRefreshToken     = errors.New("Invalid refresh token")
	ErrUnexpectedSigningMethod = errors.New("Unexpected signing method")
	ErrInvalidToken            = errors.New("Invalid token")
)
