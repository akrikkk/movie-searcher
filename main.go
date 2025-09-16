package main

import (
	"database/sql"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"movie-searcher/authorization"
	"movie-searcher/notmain"
	"movie-searcher/registration"

	_ "modernc.org/sqlite"
)

type Movie struct {
	ID          int
	Title       string
	Genre       string
	Country     string
	Year        int
	Rating      float64
	URL         sql.NullString
	Description string
	PhotoUrl    string
}

var list []Movie

func moviePageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/moviePage.html")
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		log.Printf("Ошибка шаблона: %v", err)
		return
	}
	db, err := sql.Open("sqlite", "./sql/movies.db")
	if err != nil {
		http.Error(w, "Ошибка подключения к bd", http.StatusInternalServerError)
		log.Printf("Ошибка подключения к bd: %v", err)
		return
	}
	defer db.Close()

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Некорректный ID фильма", http.StatusBadRequest)
		return
	}
	var movie Movie
	err = db.QueryRow("SELECT id, title, genre, country, year, rating, url, description, photourl FROM movies WHERE id = ?", id).Scan(&movie.ID, &movie.Title, &movie.Genre, &movie.Country, &movie.Year, &movie.Rating, &movie.URL, &movie.Description, &movie.PhotoUrl)
	if err != nil {
		http.Error(w, "Ошибка получения фильма", http.StatusInternalServerError)
		log.Printf("Ошибка получения фильма: %v", err)
		return
	}

	tmpl.Execute(w, movie)
}

func allMoviesHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/AllMovies.html")
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		return
	}

	db, err := sql.Open("sqlite", "./sql/movies.db")
	if err != nil {
		http.Error(w, "Ошибка подключения к БД", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	genre := r.URL.Query().Get("genre")
	year := r.URL.Query().Get("year")
	country := r.URL.Query().Get("country")

	query := "SELECT id, title, genre, country, year, rating, description, photourl FROM movies WHERE 1=1"
	args := []interface{}{}

	if genre != "" {
		query += " AND genre = ?"
		args = append(args, genre)
	}
	if year != "" {
		query += " AND year = ?"
		args = append(args, year)
	}
	if country != "" {
		query += " AND country = ?"
		args = append(args, country)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "Ошибка получения фильмов", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Genre, &m.Country, &m.Year, &m.Rating, &m.Description, &m.PhotoUrl); err != nil {
			continue
		}
		movies = append(movies, m)
	}

	genreRows, _ := db.Query("SELECT DISTINCT genre FROM movies")
	var genres []string
	for genreRows.Next() {
		var g string
		genreRows.Scan(&g)
		genres = append(genres, g)
	}
	genreRows.Close()

	yearRows, _ := db.Query("SELECT DISTINCT year FROM movies ORDER BY year DESC")
	var years []int
	for yearRows.Next() {
		var y int
		yearRows.Scan(&y)
		years = append(years, y)
	}
	yearRows.Close()

	countryRows, _ := db.Query("SELECT DISTINCT country FROM movies")
	var countries []string
	for countryRows.Next() {
		var c string
		countryRows.Scan(&c)
		countries = append(countries, c)
	}
	countryRows.Close()

	data := struct {
		Movies    []Movie
		Genres    []string
		Years     []int
		Countries []string
	}{
		Movies:    movies,
		Genres:    genres,
		Years:     years,
		Countries: countries,
	}

	tmpl.Execute(w, data)
}

func movieHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		log.Printf("Ошибка шаблона: %v", err)
		return
	}

	if len(list) == 0 {
		http.Error(w, "Список фильмов пуст", http.StatusInternalServerError)
		return
	}

	randomFilm := list[rand.Intn(len(list))]

	tmpl.Execute(w, randomFilm)
}

func AddMovieHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/addMovie.html")
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		log.Printf("Ошибка шаблона: %v", err)
		return
	}
	if r.Method == http.MethodGet {
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		genre := r.FormValue("genre")
		country := r.FormValue("country")
		year, _ := strconv.Atoi(r.FormValue("year"))
		rating, _ := strconv.ParseFloat(r.FormValue("rating"), 64)
		description := r.FormValue("description")
		photourl := r.FormValue("photourl")

		db, err := sql.Open("sqlite", "./sql/movies.db")
		if err != nil {
			http.Error(w, "Ошибка подключения к bd", http.StatusInternalServerError)
			log.Printf("Ошибка подключения к bd: %v", err)
			return
		}
		defer db.Close()

		_, err = db.Exec("INSERT INTO movies (title, genre, country, year, rating, description, photourl) VALUES (?, ?, ?, ?, ?, ?, ?)", title, genre, country, year, rating, description, photourl)
		if err != nil {
			http.Error(w, "Ошибка добавления фильма", http.StatusInternalServerError)
			log.Printf("Ошибка добавления фильма: %v", err)
			return
		}
		list = append(list, Movie{
			Title:       title,
			Genre:       genre,
			Country:     country,
			Year:        year,
			Rating:      rating,
			Description: description,
			PhotoUrl:    photourl,
		})

		log.Printf("Добавлен фильм: %s", title)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}

func main() {
	db, err := sql.Open("sqlite", "./sql/movies.db")
	if err != nil {
		log.Printf("Ошибка подключения к bd")
	}
	defer db.Close()
	log.Println("Подключение к bd успешно")

	rows, err := db.Query("SELECT id, title, genre, country, year, rating, url, description, photourl FROM movies")
	if err != nil {
		log.Fatalf("Ошибка получения значений из bd: %v", err)
	}

	log.Println("Данные получены из bd успешно")

	defer rows.Close()

	for rows.Next() {
		var movie Movie

		if err := rows.Scan(&movie.ID, &movie.Title, &movie.Genre, &movie.Country, &movie.Year, &movie.Rating, &movie.URL, &movie.Description, &movie.PhotoUrl); err != nil {
			log.Printf("Ошибка получения данных")
		}

		list = append(list, movie)
	}

	notmain.DB, err = sql.Open("sqlite", "./sql/movies.db")
	if err != nil {
		log.Fatal("Ошибка подключения к базе:", err)
	}

	if err := notmain.DB.Ping(); err != nil {
		log.Fatal("База недоступна:", err)
	}

	http.HandleFunc("/", authorization.AuthHandler)
	http.HandleFunc("/register", registration.RegistrationHandler)
	http.HandleFunc("/start", movieHandler)
	http.HandleFunc("/add-movie", AddMovieHandler)
	http.HandleFunc("/allMovies", allMoviesHandler)
	http.HandleFunc("/moviePage", moviePageHandler)
	http.HandleFunc("/profile", notmain.ProfilePageHandler)
	http.HandleFunc("/add-to-profile", notmain.ProfileHandler)
	http.HandleFunc("/remove-from-profile", notmain.RemoveFromProfileHandler)
	http.Handle("/sin/", http.StripPrefix("/sin/", http.FileServer(http.Dir("sin"))))

	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
