// Copyright (C) The AetherFS Authors - All Rights Reserved
//
// Proprietary and confidential.
// Unauthorized copying of this file, via any medium is strictly prohibited.
// Written by Mya Pitzeruse, September 2021

package auth

import "context"

// HandleFunc defines a generic function for handling authentication and authorization.
type HandleFunc func(ctx context.Context) (context.Context, error)
