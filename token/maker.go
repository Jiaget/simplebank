package token

import "time"

// Maker is an interface to manage the tokens
type Maker interface {
	// CreateToken creates a new token
	CreateToken(username string, duration time.Duration) (string, error)
	// VarifyToken  checks if the token valid
	VarifyToken(token string) (*PalyLoad, error)
}
