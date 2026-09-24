package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	maxUploadSize = 1 << 20 // 1 MB
	allowedPNG    = ".png"
	allowedJPG    = ".jpg"
	allowedJPEG   = ".jpeg"
)

type server struct {
	imageDir string
}

type uploadResponse struct {
	Slug string `json:"slug"`
}

func (s *server) upload(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>Upload image</title></head>
<body>
<form method="post" action="/upload" enctype="multipart/form-data">
<label>Image <input type="file" name="file" accept="image/png,image/jpeg" required></label>
<button type="submit">Upload</button>
</form>
</body>
</html>`))
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file upload is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	extension := strings.ToLower(filepath.Ext(header.Filename))
	if !isImageExtension(extension) {
		http.Error(w, "only png and jpeg images are supported", http.StatusBadRequest)
		return
	}

	if err := os.MkdirAll(s.imageDir, 0750); err != nil {
		http.Error(w, "could not create image directory", http.StatusInternalServerError)
		return
	}

	slug, destination, err := s.newImagePath(extension)
	if err != nil {
		http.Error(w, "could not create image name", http.StatusInternalServerError)
		return
	}

	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		http.Error(w, "could not save image", http.StatusInternalServerError)
		return
	}

	_, copyErr := io.Copy(output, file)
	closeErr := output.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(destination)
		http.Error(w, "could not save image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(uploadResponse{Slug: slug})
}

func (s *server) image(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(r.URL.Path, "/")
	if name == "" || name == "upload" || strings.Contains(name, "/") {
		http.NotFound(w, r)
		return
	}

	if isImageName(name) {
		s.rawImage(w, r, name)
		return
	}

	path, err := s.findImage(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, path)
}

func (s *server) rawImage(w http.ResponseWriter, r *http.Request, name string) {
	path, err := s.findImage(name)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, path)
}

func (s *server) newImagePath(extension string) (string, string, error) {
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const slugLength = 8

	for attempt := 0; attempt < 10; attempt++ {
		bytes := make([]byte, slugLength)
		if _, err := rand.Read(bytes); err != nil {
			return "", "", err
		}
		for index := range bytes {
			bytes[index] = alphabet[int(bytes[index])%len(alphabet)]
		}
		slug := string(bytes)
		path := filepath.Join(s.imageDir, slug+extension)
		if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
			return slug, path, nil
		}
	}
	return "", "", errors.New("could not find an unused image name")
}

func (s *server) findImage(name string) (string, error) {
	if filepath.Base(name) != name {
		return "", os.ErrNotExist
	}

	candidates := []string{name}
	if filepath.Ext(name) == "" {
		candidates = append(candidates, name+allowedPNG, name+allowedJPG, name+allowedJPEG)
	}
	for _, candidate := range candidates {
		if !isImageName(candidate) {
			continue
		}
		path := filepath.Join(s.imageDir, candidate)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
	}
	return "", os.ErrNotExist
}

func isImageName(name string) bool {
	return isImageExtension(strings.ToLower(filepath.Ext(name)))
}

func isImageExtension(extension string) bool {
	return extension == allowedPNG || extension == allowedJPG || extension == allowedJPEG
}

func main() {
	imageDir := os.Getenv("SNIPIT_IMAGE_DIR")
	if imageDir == "" {
		imageDir = "images"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &server{imageDir: imageDir}
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", server.upload)
	mux.HandleFunc("/", server.image)

	if err := http.ListenAndServe(":"+port, mux); err != nil {
		panic(err)
	}
}
