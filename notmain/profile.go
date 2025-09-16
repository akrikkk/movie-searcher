package notmain

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

type Movie struct {
	ID          int
	Title       string
	Genre       string
	Country     string
	Year        int
	Rating      float64
	Description string
	PhotoUrl    string
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	movieIDStr := r.FormValue("movie_id")
	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	_, err = DB.Exec("INSERT OR IGNORE INTO user_movies (user_id, movie_id) VALUES (?, ?)", 0, movieID)
	if err != nil {
		http.Error(w, "Ошибка добавления фильма", http.StatusInternalServerError)
		log.Printf("Ошибка добавления фильма: %v", err)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
	log.Println("Фильм добавлен в просмотренные")
}

func RemoveFromProfileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	movieIDStr := r.FormValue("movie_id")
	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	_, err = DB.Exec("DELETE FROM user_movies WHERE movie_id = ?", movieID)
	if err != nil {
		http.Error(w, "Ошибка удаления фильма", http.StatusInternalServerError)
		log.Printf("Ошибка удаления фильма: %v", err)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusSeeOther)
	log.Println("Фильм удалён из просмотренных")
}

func ProfilePageHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/profile.html")
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		log.Printf("Ошибка шаблона: %v", err)
		return
	}

	rows, err := DB.Query(`
		SELECT m.id, m.title, m.genre, m.country, m.year, m.rating, m.description, m.photourl
		FROM movies m
		JOIN user_movies um ON m.id = um.movie_id
		WHERE um.user_id = ?
		ORDER BY um.watched_at DESC
	`, 0)
	if err != nil {
		http.Error(w, "Ошибка получения фильмов", http.StatusInternalServerError)
		log.Printf("Ошибка получения фильмов: %v", err)
		return
	}
	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Genre, &m.Country, &m.Year, &m.Rating, &m.Description, &m.PhotoUrl); err != nil {
			log.Printf("Ошибка сканирования фильма: %v", err)
			continue
		}
		movies = append(movies, m)
	}

	data := struct {
		Movies []Movie
	}{
		Movies: movies,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Ошибка выполнения шаблона", http.StatusInternalServerError)
		log.Printf("Ошибка выполнения шаблона: %v", err)
		return
	}

	log.Println("Страница просмотренных фильмов отображена")
}
