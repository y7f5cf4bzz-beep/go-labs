package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"go-labs/mlog"
	"go-labs/mpage"
)

var dbmysqlg *sql.DB

func init() {
	mlog.LoggedUserSession.Options = &sessions.Options{
		Domain:   "localhost",
		Path:     "/",
		MaxAge:   3600 * 3,
		HttpOnly: true,
	}
}

func main() {
	host, user, password := mlog.InfoMyConn("db.txt")
	fmt.Println("Host:", host, "User:", user)

	var err error
	dbmysqlg, err = sql.Open("mysql", user+":"+password+"@tcp("+host+":3306)/activity?charset=utf8")
	mlog.CheckErr(err, "Не могу открыть БД activity")
	defer dbmysqlg.Close()

	mlog.Exdbmysqlg = dbmysqlg
	mpage.Exdbmysqlg = dbmysqlg

	// Статические файлы
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))
	http.Handle("/node_modules/", http.StripPrefix("/node_modules/", http.FileServer(http.Dir("node_modules"))))

	// Маршруты
	http.HandleFunc("/", mlog.LoginPageHandler)
	http.HandleFunc("/index", mlog.Index)
	http.HandleFunc("/logout", mlog.LogoutHandler)
	http.HandleFunc("/searchstudent", mpage.SearchStudent)
	http.HandleFunc("/searchconference", mpage.SearchConference)
	http.HandleFunc("/cityclassifier", mpage.CityClassifier)
	http.HandleFunc("/chartbasicbar", mpage.ChartBasicBar)
	http.HandleFunc("/searchreport", mpage.SearchReport)

	fmt.Println("Сервер запущен на 0.0.0.0:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		log.Fatal(err)
	}
}
