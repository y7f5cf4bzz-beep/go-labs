package mpage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"go-labs/mlog"
)

var arDC = make(map[int]map[string]int64)

func ChartBasicBar(w http.ResponseWriter, r *http.Request) {
	// Инициализация карт
	for i := 0; i <= 3; i++ {
		arDC[i] = make(map[string]int64)
	}

	// Разрешаем CORS и устанавливаем тип контента
	w.Header().Set("Content-Type", "application/json")
	
	if r.Method == "GET" {
		fmt.Fprintf(w, `{"error": "GET method not allowed"}`)
		return
	}

	// Проверяем сессию
	if !mlog.CheckSession(r) {
		fmt.Fprintf(w, `{"error": "unauthorized"}`)
		return
	}

	if err := r.ParseMultipartForm(64 << 20); err != nil {
		fmt.Println("ParseForm() err: ", err)
		fmt.Fprintf(w, `{"error": "parse form error"}`)
		return
	}

	studentid := r.FormValue("studentid")
	if studentid == "" {
		studentid = "%"
	}
	activitytype := r.FormValue("activitytype")

	switch activitytype {
	case "1":
		BasicBarJSON(studentid, 1)
		pagesJson, err := json.Marshal(arDC[1])
		if err != nil {
			fmt.Fprintf(w, `{"error": "json error"}`)
			return
		}
		fmt.Fprintf(w, "%s", pagesJson)
	case "2":
		BasicBarJSON(studentid, 2)
		pagesJson, err := json.Marshal(arDC[2])
		if err != nil {
			fmt.Fprintf(w, `{"error": "json error"}`)
			return
		}
		fmt.Fprintf(w, "%s", pagesJson)
	case "3":
		BasicBarJSON(studentid, 3)
		pagesJson, err := json.Marshal(arDC[3])
		if err != nil {
			fmt.Fprintf(w, `{"error": "json error"}`)
			return
		}
		fmt.Fprintf(w, "%s", pagesJson)
	default:
		var wg sync.WaitGroup
		for i := 1; i <= 3; i++ {
			wg.Add(1)
			go func(i int) {
				BasicBarJSON(studentid, i)
				wg.Done()
			}(i)
		}
		wg.Wait()
		for k := range arDC[1] {
			arDC[0][k] = arDC[1][k] + arDC[2][k] + arDC[3][k]
		}
		pagesJson, err := json.Marshal(arDC[0])
		if err != nil {
			fmt.Fprintf(w, `{"error": "json error"}`)
			return
		}
		fmt.Fprintf(w, "%s", pagesJson)
	}
}

func BasicBarJSON(studentid string, activitytype int) {
	var sSQLPoint string

	switch activitytype {
	case 1:
		sSQLPoint = `IFNULL((SELECT SUM(student_conference.point) FROM student_conference 
			LEFT JOIN conference ON student_conference.conference_id=conference.id 
			WHERE student_conference.student_id=student.id), 0)`
	case 2:
		sSQLPoint = `IFNULL((SELECT SUM(IFNULL(student_project.point,0)) FROM student_project 
			LEFT JOIN project ON student_project.project_id=project.id 
			WHERE student_project.student_id=student.id), 0)`
	case 3:
		sSQLPoint = `IFNULL((SELECT SUM(IFNULL(student_paper.point,0)) FROM student_paper 
			LEFT JOIN paper ON student_paper.paper_id=paper.id 
			WHERE student_paper.student_id=student.id), 0)`
	}

	stmt, err := Exdbmysqlg.Prepare(`SELECT student.fio, ` + sSQLPoint + ` as std_point 
		FROM student 
		WHERE student.id LIKE ? 
		ORDER BY fio`)
	mlog.CheckErr(err, "Не могу подготовить запрос")
	defer stmt.Close()

	rows, err := stmt.Query(studentid)
	mlog.CheckErr(err, "Не могу выполнить запрос")
	defer rows.Close()

	var fio sql.NullString
	var std_point sql.NullInt64

	for rows.Next() {
		err = rows.Scan(&fio, &std_point)
		mlog.CheckErr(err, "Не могу прочесть запись")
		arDC[activitytype][fio.String] = std_point.Int64
	}
}
