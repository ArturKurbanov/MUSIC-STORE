package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Создаем альбом
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

var storage = NewStorage()

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

	album, err := storage.Update(id, newAlbum)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"massage": "album not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, album)
}

func getRouter() *gin.Engine {
	router := gin.Default()
	gin.SetMode(gin.ReleaseMode)
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
