package authorization

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

var User struct {
	ID       int
	Username string
	Password string
}

func HashPassword(password string) (string, error) {
	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedpassword), nil

}
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false
	}
	return true

}

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("./templates/auth.html")
	if err != nil {
		http.Error(w, "Ошибка шаблона", http.StatusInternalServerError)
		log.Printf("Ошибка шаблона: %v", err)
		return
	}
	if r.Method == http.MethodGet {
		tmpl.Execute(w, nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password_hash")

	db, err := sql.Open("sqlite", "./sql/movies.db")
	if err != nil {
		http.Error(w, "Ошибка подключения к bd", http.StatusInternalServerError)
		log.Printf("Ошибка подключения к bd: %v", err)
		return
	}
	defer db.Close()
	log.Println("Подключение к базе данных пользователей успешно")

	row := db.QueryRow("SELECT id, username, password_hash FROM users WHERE username = ?", username)
	err = row.Scan(&User.ID, &User.Username, &User.Password)
	if err != nil {
		http.Error(w, "пользователь не найден", http.StatusInternalServerError)
		log.Printf("Пользователь не найден: %v", err)
		return
	}
	if !CheckPassword(password, User.Password) {
		http.Error(w, "неверный пароль", http.StatusUnauthorized)
		log.Printf("неверный пароль: %v", err)
		return
	}
	log.Println("Авторизация прошла успешно")
	http.Redirect(w, r, "/start", http.StatusSeeOther)
}
