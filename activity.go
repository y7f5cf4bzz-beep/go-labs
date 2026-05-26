package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"go-labs/mlog"
)

var dbmysqlg *sql.DB
var err error

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

	dbmysqlg, err = sql.Open("mysql", user+":"+password+"@tcp("+host+":3306)/activity?charset=utf8")
	mlog.CheckErr(err, "Не могу открыть БД activity")
	defer dbmysqlg.Close()

	fmt.Printf("Тип dbmysqlg: %T\n", dbmysqlg)

	mlog.Exdbmysqlg = dbmysqlg

	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("public"))))

	http.HandleFunc("/", mlog.LoginPageHandler)
	http.HandleFunc("/index", mlog.Index)
	http.HandleFunc("/logout", mlog.LogoutHandler)

	fmt.Println("Сервер запущен на :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
