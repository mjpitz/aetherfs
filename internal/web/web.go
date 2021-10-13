package web

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist/*
var assets embed.FS

// Handle must be mounted on /ui/.
func Handle() http.Handler {
	fs, err := fs.Sub(fs.FS(assets), "dist")
	if err != nil {
		panic(err)
	}
	return http.StripPrefix("/ui/", http.FileServer(http.FS(fs)))
}
