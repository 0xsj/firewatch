package oauth

import "errors"

var (
	// Configuration errors
	ErrMissingClientID     = errors.New("oauth: client ID is required")
	ErrMissingClientSecret = errors.New("oauth: client secret is required")
	ErrMissingRedirectURL  = errors.New("oauth: redirect URL is required")

	// Flow errors
	ErrInvalidState        = errors.New("oauth: invalid state token")
	ErrExpiredState        = errors.New("oauth: state token expired")
	ErrInvalidCode         = errors.New("oauth: invalid authorization code")
	ErrTokenExchange       = errors.New("oauth: failed to exchange token")
	ErrUserInfoFetch       = errors.New("oauth: failed to fetch user info")
	ErrRefreshTokenInvalid = errors.New("oauth: refresh token is invalid")
)
