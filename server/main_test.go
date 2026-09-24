package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUploadAndServeImage(t *testing.T) {
	imageDir := t.TempDir()
	server := &server{imageDir: imageDir}
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", server.upload)
	mux.HandleFunc("/", server.image)

	formRequest := httptest.NewRequest(http.MethodGet, "/upload", nil)
	formResponse := httptest.NewRecorder()
	mux.ServeHTTP(formResponse, formRequest)
	if formResponse.Code != http.StatusOK {
		t.Fatalf("upload form status = %d, want %d", formResponse.Code, http.StatusOK)
	}
	if contentType := formResponse.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Fatalf("upload form content type = %q, want text/html; charset=utf-8", contentType)
	}
	form := formResponse.Body.String()
	for _, expected := range []string{`method="post"`, `action="/upload"`, `enctype="multipart/form-data"`, `name="file"`} {
		if !strings.Contains(form, expected) {
			t.Fatalf("upload form does not contain %q", expected)
		}
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "capture.png")
	if err != nil {
		t.Fatal(err)
	}
	content := []byte("fake png content")
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	uploadRequest := httptest.NewRequest(http.MethodPost, "/upload?redirect=0", &body)
	uploadRequest.Header.Set("Content-Type", writer.FormDataContentType())
	uploadRecorder := httptest.NewRecorder()
	mux.ServeHTTP(uploadRecorder, uploadRequest)
	if uploadRecorder.Code != http.StatusOK {
		t.Fatalf("upload status = %d, want %d", uploadRecorder.Code, http.StatusOK)
	}

	var result uploadResponse
	if err := json.Unmarshal(uploadRecorder.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(result.Slug) != "" || len(result.Slug) != 8 {
		t.Fatalf("slug = %q, want an eight-character extensionless ID", result.Slug)
	}

	imageRequest := httptest.NewRequest(http.MethodGet, "/"+result.Slug, nil)
	imageResponse := httptest.NewRecorder()
	mux.ServeHTTP(imageResponse, imageRequest)
	if imageResponse.Code != http.StatusOK || !bytes.Equal(imageResponse.Body.Bytes(), content) {
		t.Fatalf("served image = (%d, %q), want (%d, %q)", imageResponse.Code, imageResponse.Body.Bytes(), http.StatusOK, content)
	}

	rawRequest := httptest.NewRequest(http.MethodGet, "/"+result.Slug+".png", nil)
	rawResponse := httptest.NewRecorder()
	mux.ServeHTTP(rawResponse, rawRequest)
	if rawResponse.Code != http.StatusOK || !bytes.Equal(rawResponse.Body.Bytes(), content) {
		t.Fatalf("raw image = (%d, %q), want (%d, %q)", rawResponse.Code, rawResponse.Body.Bytes(), http.StatusOK, content)
	}

	lookupRequest := httptest.NewRequest(http.MethodGet, "/"+result.Slug, nil)
	lookupResponse := httptest.NewRecorder()
	mux.ServeHTTP(lookupResponse, lookupRequest)
	if lookupResponse.Code != http.StatusOK || !bytes.Equal(lookupResponse.Body.Bytes(), content) {
		t.Fatalf("extensionless image = (%d, %q), want (%d, %q)", lookupResponse.Code, lookupResponse.Body.Bytes(), http.StatusOK, content)
	}

	entries, err := os.ReadDir(imageDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("stored files = %d, want 1", len(entries))
	}
}
