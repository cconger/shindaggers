package main

import (
	"embed"
	"io"
	"log"
	"net/http"
	"path/filepath"
)

//go:embed client/index.html client/assets
var assets embed.FS

func (s *Server) assetHandler(w http.ResponseWriter, r *http.Request) {
	p, err := filepath.Abs(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	assetPath := filepath.Join("client", p)

	// Return the path of the asset
	f, err := assets.Open(assetPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer f.Close()

	ext := filepath.Ext(assetPath)
	switch ext {
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	}

	_, err = io.Copy(w, f)
	if err != nil {
		log.Printf("Error writing file: %s", err.Error())
	}
}

func (s *Server) spaHandler(w http.ResponseWriter, r *http.Request) {
	// Always return index.html
	f, err := assets.Open("client/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = io.Copy(w, f)
	if err != nil {
		log.Printf("Error writing file: %s", err.Error())
	}
}
