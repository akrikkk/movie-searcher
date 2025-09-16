package registration

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/registration.html")
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

	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	_, err = db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, string(hashed))
	if err != nil {
		http.Error(w, "Неверные данные", http.StatusInternalServerError)
		log.Printf("Ошибка регистрации пользователя: %v", err)
		return
	}
	log.Println("Регистрация прошла успешно")
	http.Redirect(w, r, "/auth", http.StatusSeeOther)
}
