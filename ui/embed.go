package ui

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
)

//go:embed build/*
var buildFS embed.FS

func Handler() http.HandlerFunc {
	subFS, err := fs.Sub(buildFS, "build")
	if err != nil {
		log.Fatalf("failed to create sub filesystem: %v", err)
	}
	handler := http.FileServer(http.FS(subFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) == "" {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
			r.URL.Path = "/"
		}
		handler.ServeHTTP(w, r)
	})
}
