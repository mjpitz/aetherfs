package auth

import "context"

// HandleFunc defines a generic function for handling authentication and authorization.
type HandleFunc func(ctx context.Context) (context.Context, error)
