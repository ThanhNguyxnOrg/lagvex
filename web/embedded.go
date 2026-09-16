package web

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
)

//go:embed index.html style.css app.js assets/* assets/games/*
var EmbeddedFS embed.FS

// GetFileSystem returns http.FileSystem prioritizing disk (if exists) for live editing,
// otherwise falling back to the embedded files in the binary.
func GetFileSystem(diskDir string) http.FileSystem {
	if diskDir != "" {
		if fi, err := os.Stat(diskDir); err == nil && fi.IsDir() {
			return http.Dir(diskDir)
		}
	}
	// Fallback to embedded filesystem
	sub, err := fs.Sub(EmbeddedFS, ".")
	if err == nil {
		return http.FS(sub)
	}
	return http.FS(EmbeddedFS)
}
