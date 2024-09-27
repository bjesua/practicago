package main

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	var err error

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", dbUser, dbPassword, dbHost, dbName)

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
		secret := os.Getenv("SECRET_KEY")
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

	router.Run(":8081")
}

func fetchDataFromURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

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
	song := c.Query("song")
	artist := c.Query("artist")
	album := c.Query("album")

	// Construir la URL de búsqueda
	url := os.Getenv("PATH_APPLE") + url.QueryEscape(song) + "+" + url.QueryEscape(artist) + "+" + url.QueryEscape(album)

	data, err := fetchDataFromURL(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var originalData OriginalData
	if err := json.Unmarshal([]byte(data), &originalData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

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

	// Llamada a ChartLyrics
	if song != "" && artist != "" {
		lyricsData, err := getDataChartLyrics(artist, song)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for _, result := range lyricsData {
			newData = append(newData, NewData{
				ID:       result.ID,
				Name:     result.Name,
				Artist:   result.Artist,
				Album:    result.Album,
				Artwork:  result.Artwork,
				Price:    result.Price,
				Origin:   result.Origin,
				Duration: result.Duration,
			})
		}
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

	newData, err = show_results(song, artist, album)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, newData)
}

func CreateSong(song NewData) error {
	query := "INSERT INTO songs (name, artist, album, artwork, duration, price, origin) VALUES (?, ?, ?, ?, ?, ?, ?)"
	_, err := db.Exec(query, song.Name, song.Artist, song.Album, song.Artwork, song.Duration, song.Price, song.Origin)
	return err
}

// Estructura para parsear el XML de ChartLyrics
type ArrayOfSearchLyricResult struct {
	SearchLyricResults []SearchLyricResult `xml:"SearchLyricResult"`
}

type SearchLyricResult struct {
	TrackId       int    `xml:"TrackId"`
	LyricChecksum string `xml:"LyricChecksum"`
	LyricId       int    `xml:"LyricId"`
	SongUrl       string `xml:"SongUrl"`
	ArtistUrl     string `xml:"ArtistUrl"`
	Artist        string `xml:"Artist"`
	Song          string `xml:"Song"`
	SongRank      int    `xml:"SongRank"`
}

func getDataChartLyrics(artist string, song string) ([]NewData, error) {

	searchURL := os.Getenv("CHARTLYRICS_API") + url.QueryEscape(artist) + "&song=" + url.QueryEscape(song)

	data, err := fetchDataFromURL(searchURL)
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %v", err)
	}

	var lyricsData ArrayOfSearchLyricResult
	err = xml.Unmarshal([]byte(data), &lyricsData)
	if err != nil {
		return nil, fmt.Errorf("error parsing XML: %v", err)
	}

	var newData []NewData
	for _, result := range lyricsData.SearchLyricResults {
		newData = append(newData, NewData{
			ID:       result.TrackId,
			Name:     result.Song,
			Artist:   result.Artist,
			Album:    "",
			Artwork:  result.ArtistUrl,
			Price:    0,
			Origin:   "ChartLyrics",
			Duration: "0:00",
		})
	}

	return newData, nil
}

func show_results(song string, artist string, album string) ([]NewData, error) {
	// Inicializar la consulta base
	query := "SELECT id, name, artist, album, artwork, price, origin, duration FROM songs WHERE id <> ''"

	// var params []interface{}

	if song != "" {
		query += " or name = '%" + song + "%'"
		// params = append(params, song)
	}
	if artist != "" {
		query += " or artist = '%" + artist + "%'"
		// params = append(params, artist)
	}
	if album != "" {
		query += " or album = '%" + album + "%'"
		// params = append(params, album)
	}

	fmt.Println(query)

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %v", err)
	}
	defer rows.Close()

	var databasedata []NewData

	for rows.Next() {
		var song NewData

		err = rows.Scan(&song.ID, &song.Name, &song.Artist, &song.Album, &song.Artwork, &song.Price, &song.Origin, &song.Duration)
		if err != nil {
			return nil, fmt.Errorf("error scanning data: %v", err)
		}

		databasedata = append(databasedata, song)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %v", err)
	}

	return databasedata, nil
}
