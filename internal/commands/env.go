package commands

import "os"

func lookupHost() string {
	host := os.Getenv("AETHERFS_HOST")
	if host != "" {
		return host
	}
	return "localhost:8080"
}
