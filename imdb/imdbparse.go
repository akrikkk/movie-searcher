package 
import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
)

type TMDbResponse struct {
	Page    int         `json:"page"`
	Results []TMDbMovie `json:"results"`
}

type TMDbMovie struct {
	Title       string  `json:"title"`
	Overview    string  `json:"overview"`
	ReleaseDate string  `json:"release_date"`
	VoteAverage float64 `json:"vote_average"`
	PosterPath  string  `json:"poster_path"`
}

func fetchMovies(apiKey string, page int) ([]TMDbMovie, error) {
	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/popular?api_key=%s&language=ru-RU&page=%d", apiKey, page)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tmdbResp TMDbResponse
	err = json.Unmarshal(body, &tmdbResp)
	if err != nil {
		return nil, err
	}

	return tmdbResp.Results, nil
}

func saveMovie(db *sql.DB, movie TMDbMovie) error {
	posterURL := "https://image.tmdb.org/t/p/w500" + movie.PosterPath
	year := 0
	if len(movie.ReleaseDate) >= 4 {
		year, _ = strconv.Atoi(movie.ReleaseDate[:4])
	}

	_, err := db.Exec(`INSERT INTO movies (title, genre, country, year, rating, description, photourl)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		movie.Title, "Unknown", "Unknown", year, movie.VoteAverage, movie.Overview, posterURL)

	return err
}

func main() {
	apiKey := "d7d30f2765ec4dbb49de0b01a8d4fb05"
	db, err := sql.Open("sqlite", "C:/Users/akiri/yanprojects/sql/movies.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	for page := 56; page <= 65; page++ {
		movies, err := fetchMovies(apiKey, page)
		if err != nil {
			log.Println("Ошибка получения фильмов:", err)
			continue
		}

		for _, movie := range movies {
			err := saveMovie(db, movie)
			if err != nil {
				log.Println("Ошибка сохранения фильма:", err)
			} else {
				log.Println("Сохранён фильм:", movie.Title)
			}
		}
	}
}
