package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// router serves index.html for any route that is not a static file
func router(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/static/") {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, "./static/index.html")
}

// serveStatic serves JS/CSS/JSON files with correct MIME types
func serveStatic(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(".", r.URL.Path)

	// Open the file
	f, err := os.Open(path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	// Determine MIME type manually
	switch {
	case strings.HasSuffix(r.URL.Path, ".js"):
		w.Header().Set("Content-Type", "application/javascript")
	case strings.HasSuffix(r.URL.Path, ".css"):
		w.Header().Set("Content-Type", "text/css")
	case strings.HasSuffix(r.URL.Path, ".json"):
		w.Header().Set("Content-Type", "application/json")
	default:
		w.Header().Set("Content-Type", "text/plain")
	}

	// Serve file content manually so our Content-Type is respected
	http.ServeContent(w, r, path, time.Now(), f)
}

func main() {
	http.HandleFunc("/static/", serveStatic)
	http.HandleFunc("/", router)

	log.Println("Frontend server running on http://localhost:8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal(err)
	}
}
