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

	"github.com/gin-gonic/gin"
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
	url := "https://itunes.apple.com/search?term=" + url.QueryEscape(song) + "+" + url.QueryEscape(artist) + "+" + url.QueryEscape(album)

	// Llamar a la función que obtiene los datos de ChartLyrics
	lyricsData, err := getDataChartLyrics(artist, song)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
	searchURL := "http://api.chartlyrics.com/apiv1.asmx/SearchLyric?artist=" + url.QueryEscape(artist) + "&song=" + url.QueryEscape(song)

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
			Album:    "", // No hay información sobre el álbum
			Artwork:  result.ArtistUrl,
			Price:    0,
			Origin:   "ChartLyrics",
			Duration: "0:00",
		})
	}

	return newData, nil
}

func show_results(song, artist, album string) ([]NewData, error) {
	// Inicializar la consulta base
	query := "SELECT id, name, artist, album, artwork, price, origin, duration FROM songs"

	// Utilizar un slice para agregar condiciones de forma dinámica
	// var conditions []string
	// var params []interface{}

	// // Construir dinámicamente el WHERE dependiendo de los valores recibidos
	// if song != "" {
	// 	conditions = append(conditions, "name = ?")
	// 	params = append(params, song)
	// }
	// if artist != "" {
	// 	conditions = append(conditions, "artist = ?")
	// 	params = append(params, artist)
	// }
	// if album != "" {
	// 	conditions = append(conditions, "album = ?")
	// 	params = append(params, album)
	// }

	// // Si hay condiciones, agregarlas a la consulta
	// if len(conditions) > 0 {
	// 	query = fmt.Sprintf("%s WHERE %s", query, strings.Join(conditions, " AND "))
	// }

	// Imprimir la consulta y los parámetros para depuración
	// fmt.Println("Generated Query:", query)
	// fmt.Println("Query Parameters:", params)
	// return []NewData{}, nil
	// Ejecutar la consulta con los parámetros
	rows, err := db.Query(query)
	// rows, err := db.Query(query, params...)
	if err != nil {
		return nil, fmt.Errorf("error fetching data: %v", err)
	}
	defer rows.Close()

	// Slice para almacenar los resultados
	var databasedata []NewData

	// Recorrer las filas devueltas por la consulta
	for rows.Next() {
		var song NewData

		// Escanear cada fila y asignar a la estructura song
		err = rows.Scan(&song.ID, &song.Name, &song.Artist, &song.Album, &song.Artwork, &song.Price, &song.Origin, &song.Duration)
		if err != nil {
			return nil, fmt.Errorf("error scanning data: %v", err)
		}

		// Añadir la canción a la lista de resultados
		databasedata = append(databasedata, song)
	}

	// Comprobar si hubo algún error al recorrer las filas
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %v", err)
	}

	// Retornar los resultados
	return databasedata, nil
}
