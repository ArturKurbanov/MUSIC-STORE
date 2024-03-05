package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAlbumsList(t *testing.T) {
	router := getRouter()
	request, _ := http.NewRequest("GET", "/albums", strings.NewReader(""))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)
	if w.Code != http.StatusOK {
		t.Fatal("status not okg")
	}
}

func TestAlbumsDetail(t *testing.T) {
	router := getRouter()
	albumId := "1"
	request, _ := http.NewRequest("GET", "/albums/"+albumId, strings.NewReader(""))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)
	if w.Code != http.StatusOK {
		t.Fatal("status not okg")
	}
}

func TestAlbumsNotFound(t *testing.T) {
	router := getRouter()
	albumId := "9999"
	request, _ := http.NewRequest("GET", "/albums/"+albumId, strings.NewReader(""))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, request)
	if w.Code != http.StatusNotFound {
		t.Fatal("status must be 404")
	}
}
