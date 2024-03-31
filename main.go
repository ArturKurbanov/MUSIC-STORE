package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Storage interface {
	Create() album
	Read() album
	ReadOne() album
	Upgate() album
	Delete() album
}

// Создаем альбом
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

type MemoryStorage struct {
	albums []album
}

var storage = NewMemoryStorage()

func (s MemoryStorage) Create(am album) album {
	s.albums = append(s.albums, am)
	return am
}

func (s MemoryStorage) ReadOne(id string) (album, error) {
	for _, a := range s.albums {
		if a.ID == id {
			return a, nil
		}
	}
	return album{}, errors.New("album not found")
}

func (s MemoryStorage) Read() []album {
	return s.albums
}

func (s MemoryStorage) Upgate(id string, newAlbum album) (album, error) {
	for i := range s.albums {
		if s.albums[i].ID == id {
			s.albums[i] = newAlbum
			return s.albums[i], nil
		}
	}
	return album{}, errors.New("album not found")
}

func (s MemoryStorage) Delete(id string) error {
	for i, a := range s.albums {
		if a.ID == id {
			s.albums = append(s.albums[:i], s.albums[i+1:]...)
			return nil
		}
	}
	return errors.New("album not found")
}

func NewMemoryStorage() MemoryStorage {
	var albums = []album{
		{ID: "1", Title: "Blue Train", Artist: "John Coltraine", Price: 56.99},
		{ID: "2", Title: "Jeru", Artist: "Gerry Mullingan", Price: 17.99},
		{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
	}
	return MemoryStorage{albums: albums}
}

// type HttpError struct {
// 	Error string `json:"error"`
// }

// Заполняем альбом
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltraine", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mullingan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, storage.Read())
}

func postAlbums(c *gin.Context) {
	var newAlbum album
	if err := c.BindJSON(&newAlbum); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"massage": "album not found"})
		return
	}
	storage.Create(newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

func getAlbumById(c *gin.Context) {
	id := c.Param("id")
	album, err := storage.ReadOne(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"massage": "album not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, album)
}

func deleteAlbumById(c *gin.Context) {
	id := c.Param("id")
	err := storage.Delete(id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"massage": "album not found"})
		return
	}
	c.IndentedJSON(http.StatusNoContent, album{})
}

func updateAlbumById(c *gin.Context) {
	id := c.Param("id")
	var newAlbum album
	c.BindJSON(&newAlbum)

	album, err := storage.Upgate(id, newAlbum)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"massage": "album not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, album)
}

func getRouter() *gin.Engine { // gin.Engine - это основной компонент фреймворка Gin, который представляет собой маршрутизатор (router) для обработки HTTP-запросов и управления маршрутами (routes) веб-приложения.
	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumById)
	router.DELETE("/albums/:id", deleteAlbumById)
	router.PUT("/albums/:id", updateAlbumById)
	router.POST("/albums", postAlbums)
	return router
}

func main() {
	router := getRouter()
	router.Run("localhost:8080")
}
