package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	_ "modernc.org/sqlite"
)

type SearchResponse struct {
	Results []struct {
		ID int `json:"id"`
	} `json:"results"`
}

type MovieDetails struct {
	Genres []struct {
		Name string `json:"name"`
	} `json:"genres"`
	ProductionCountries []struct {
		Name string `json:"name"`
	} `json:"production_countries"`
}

func searchMovie(apiKey, title string) (int, error) {
	url := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s&language=ru-RU&query=%s",
		apiKey, url.QueryEscape(title))

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var searchResp SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return 0, err
	}

	if len(searchResp.Results) == 0 {
		return 0, fmt.Errorf("фильм не найден")
	}

	return searchResp.Results[0].ID, nil
}

func fetchMovieDetails(apiKey string, tmdbID int) (*MovieDetails, error) {
	url := fmt.Sprintf("https://api.themoviedb.org/3/movie/%d?api_key=%s&language=ru-RU", tmdbID, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var details MovieDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, err
	}

	return &details, nil
}

func main() {
	apiKey := "d7d30f2765ec4dbb49de0b01a8d4fb05"
	db, err := sql.Open("sqlite", "C:/Users/akiri/yanprojects/sql/movies.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, _ = db.Exec("PRAGMA journal_mode = WAL;")
	_, _ = db.Exec("PRAGMA synchronous = NORMAL;")

	rows, err := db.Query(`SELECT id, title FROM movies WHERE genre='Unknown' OR country='Unknown'`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var title string
		rows.Scan(&id, &title)

		tmdbID, err := searchMovie(apiKey, title)
		if err != nil {
			log.Println("Не нашли фильм:", title, err)
			continue
		}

		details, err := fetchMovieDetails(apiKey, tmdbID)
		if err != nil {
			log.Println("Ошибка деталей:", title, err)
			continue
		}

		// жанры
		genre := "Unknown"
		if len(details.Genres) > 0 {
			genre = details.Genres[0].Name
		}

		// страны
		country := "Unknown"
		if len(details.ProductionCountries) > 0 {
			country = details.ProductionCountries[0].Name
		}

		_, err = db.Exec(`UPDATE movies SET genre=?, country=? WHERE id=?`, genre, country, id)
		if err != nil {
			log.Println("Ошибка обновления:", title, err)
		} else {
			log.Println("Обновили:", title, "->", genre, "/", country)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
