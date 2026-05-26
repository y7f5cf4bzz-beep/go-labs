package mlog

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var encryptionKey = "13OtdSecret"
var LoggedUserSession = sessions.NewCookieStore([]byte(encryptionKey))
var ExSignLogin = 0
var Exdbmysqlg *sql.DB
var sGDisplayName = ""

type Page struct {
	Date        string
	Username    string
	Displayname string
}

var logUserTemplate = template.Must(template.New("").Parse(MyReadFile("public/html/login.html")))

func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	conditionsMap := map[string]interface{}{}

	if r.FormValue("Login") != "" && r.FormValue("Username") != "" {
		username := r.FormValue("Username")
		password := r.FormValue("Password")

		var sUP = ""
		sUP, sGDisplayName = GetPassDisplayName(fmt.Sprintf("%v", username))
		hashedPasswordFromDatabase := []byte(sUP)

		if err := bcrypt.CompareHashAndPassword(hashedPasswordFromDatabase, []byte(password)); err != nil {
			log.Println("Неверный логин или пароль")
			conditionsMap["LoginError"] = true
		} else {
			log.Println("Успешный вход:", username)
			conditionsMap["Username"] = username
			conditionsMap["LoginError"] = false
			session, _ := LoggedUserSession.New(r, "my-user-session")
			session.Values["username"] = username
			ExSignLogin = 1
			session.Save(r, w)
			http.Redirect(w, r, "/index", http.StatusFound)
			return
		}
	}

	if err := logUserTemplate.Execute(w, conditionsMap); err != nil {
		log.Println(err)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := LoggedUserSession.Get(r, "my-user-session")
	session.Values["username"] = ""
	session.Save(r, w)
	http.Redirect(w, r, "/", http.StatusFound)
}

func GetPassDisplayName(un string) (string, string) {
	stmt, err := Exdbmysqlg.Prepare("SELECT user_pass, full_name FROM users WHERE user_name=?;")
	CheckErr(err, "Не могу подготовить запрос")
	defer stmt.Close()

	res, err := stmt.Query(un)
	CheckErr(err, "Не могу выполнить запрос")
	defer res.Close()

	var user_pass, user_displayname sql.NullString
	for res.Next() {
		err = res.Scan(&user_pass, &user_displayname)
		CheckErr(err, "Не могу прочесть запись")
		break
	}
	return user_pass.String, user_displayname.String
}

func CheckLoginGET(w http.ResponseWriter, r *http.Request) {
	session, err := LoggedUserSession.Get(r, "my-user-session")
	if err != nil {
		log.Println("Ошибка получения сессии:", err)
	}
	if session.Values["username"] == "" || ExSignLogin == 0 {
		http.Redirect(w, r, "/logout", http.StatusFound)
	}
}

func CheckLoginPOST(w http.ResponseWriter, r *http.Request) int {
	session, err := LoggedUserSession.Get(r, "my-user-session")
	if err != nil {
		log.Println("Ошибка получения сессии:", err)
	}
	if session.Values["username"] == "" || ExSignLogin == 0 {
		return 0
	}
	return 1
}

func Index(w http.ResponseWriter, r *http.Request) {
	session, err := LoggedUserSession.Get(r, "my-user-session")
	if err != nil {
		log.Println("Ошибка получения сессии:", err)
	}

	if session.Values["username"] == "" || sGDisplayName == "" {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	year, month, day := time.Now().Date()
	curdate := fmt.Sprintf("%02d.%02d.%d", day, month, year)
	username := fmt.Sprintf("%v", session.Values["username"])

	p := &Page{
		Date:        curdate,
		Username:    username,
		Displayname: sGDisplayName,
	}

	t := template.Must(template.ParseFiles("public/html/index.html"))
	t.Execute(w, p)
}

func DateToRus(date string) string {
	date = strings.TrimSpace(date)
	if date == "" {
		return date
	}
	ar := strings.Split(date, "-")
	if len(ar) < 3 {
		return date
	}
	return ar[2] + "." + ar[1] + "." + ar[0]
}
