package static

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
)

//go:embed *
var files embed.FS

// MustGetFile returns the contents of a file from static as bytes.
func MustGetFile(name string) []byte {
	b, err := files.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return b
}

// GetFilesystem returns a http.FileSystem for the static files.
func GetFilesystem() http.FileSystem {
	return http.FS(files)
}

// GetSubFilesystem returns a http.FileSystem for the static sub-files.
func GetSubFilesystem(name string) http.FileSystem {
	fsys, err := fs.Sub(files, name)
	if err != nil {
		log.Fatalf("error loading sub-filesystem for %q: %s", name, err)
	}
	return http.FS(fsys)
}
