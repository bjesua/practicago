package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	// "api/controllers"

	"database/sql"

	"net/url"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	var err error
	dsn := "root:root@tcp(127.0.0.1:3306)/database"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Println("Conexión a la base de datos MySQL exitosa.")
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("API-Key")
		secret := "Tribal_Token"
		// secret := os.Getenv("SECRET_KEY")
		if apiKey != secret {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})
	router.Use(AuthMiddleware())

	router.GET("/api", ExternalData)

	router.Run(":8080")
}

func fetchDataFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Leer el cuerpo de la respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

type OriginalData struct {
	Results []struct {
		TrackId          int     `json:"trackId"`
		TrackName        string  `json:"trackName"`
		ArtistName       string  `json:"artistName"`
		CollectionName   string  `json:"collectionName"`
		ArtworkUrl100    string  `json:"artworkUrl100"`
		TrackRentalPrice float64 `json:"trackRentalPrice"`
	} `json:"results"`
}

type NewData struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Artist   string  `json:"artist"`
	Album    string  `json:"album"`
	Artwork  string  `json:"artwork"`
	Price    float64 `json:"price"`
	Origin   string  `json:"origin"`
	Duration string  `json:"duration"`
}

func ExternalData(c *gin.Context) {

	nombre_cancion := c.Query("nombre_cancion")
	artista := c.Query("artista")
	album := c.Query("album")

	valores := url.QueryEscape(nombre_cancion + " " + artista + " " + album)

	url := "https://itunes.apple.com/search?term=" + valores
	// url := os.Getenv("PATH_APPLE") + valores

	// url := "http://api.chartlyrics.com/apiv1.asmx/SearchLyric?artist=michael&song=thriller"

	data, err := fetchDataFromURL(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	fmt.Print(data)

	var originalData OriginalData
	if err := json.Unmarshal([]byte(data), &originalData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	fmt.Println(originalData)

	var newData []NewData
	for _, result := range originalData.Results {
		newData = append(newData, NewData{
			ID:       result.TrackId,
			Name:     result.TrackName,
			Artist:   result.ArtistName,
			Album:    result.CollectionName,
			Artwork:  result.ArtworkUrl100,
			Price:    result.TrackRentalPrice,
			Origin:   "Apple",
			Duration: "0:00",
		})
	}

	newDataJSON, err := json.Marshal(newData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	for _, song := range newData {
		err := CreateSong(song)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	c.Data(http.StatusOK, "application/json", newDataJSON)
}

func CreateSong(song NewData) error {
	query := "INSERT INTO songs (name, artist, album, artwork, duration, price, origin) VALUES (?, ?, ?, ?, ?, ?, ?)"
	_, err := db.Exec(query, song.Name, song.Artist, song.Album, song.Artwork, song.Duration, song.Price, song.Origin)
	if err != nil {
		return err
	}
	return nil
}
