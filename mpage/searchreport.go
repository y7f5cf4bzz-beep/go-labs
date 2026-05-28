package mpage

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
	"github.com/nguyenthenguyen/docx"
	"go-labs/mlog"
)

func SearchReport(w http.ResponseWriter, r *http.Request) {
	// Проверяем сессию для всех типов запросов
	if !mlog.CheckSession(r) {
		http.Redirect(w, r, "/logout", http.StatusFound)
		return
	}

	format := ""
	studentid := ""
	confname := ""
	projectname := ""
	papername := ""

	if r.Method == "GET" {
		query := r.URL.Query()
		format = query.Get("format")
		studentid = query.Get("studentid")
		if studentid == "" {
			studentid = "%"
		}
		confname = query.Get("confname")
		if confname == "" {
			confname = "%"
		}
		projectname = query.Get("projectname")
		if projectname == "" {
			projectname = "%"
		}
		papername = query.Get("papername")
		if papername == "" {
			papername = "%"
		}
	} else {
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			fmt.Println("ParseForm() err: ", err)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		format = r.FormValue("format")
		studentid = r.FormValue("studentid")
		if studentid == "" {
			studentid = "%"
		}
		confname = r.FormValue("confname")
		if confname == "" {
			confname = "%"
		}
		projectname = r.FormValue("projectname")
		if projectname == "" {
			projectname = "%"
		}
		papername = r.FormValue("papername")
		if papername == "" {
			papername = "%"
		}
	}

	stmt, err := Exdbmysqlg.Prepare(`SELECT student.fio, 
		IFNULL((SELECT SUM(student_conference.point) FROM student_conference 
			LEFT JOIN conference ON student_conference.conference_id=conference.id 
			WHERE student_conference.student_id=student.id AND conference.name LIKE ?), 0) as conference_point,
		IFNULL((SELECT SUM(IFNULL(student_project.point,0)) FROM student_project 
			LEFT JOIN project ON student_project.project_id=project.id 
			WHERE student_project.student_id=student.id AND project.name LIKE ?), 0) as project_point,
		IFNULL((SELECT SUM(IFNULL(student_paper.point,0)) FROM student_paper 
			LEFT JOIN paper ON student_paper.paper_id=paper.id 
			WHERE student_paper.student_id=student.id AND paper.name LIKE ?), 0) as paper_point
		FROM student 
		WHERE student.id LIKE ? 
		ORDER BY fio`)
	mlog.CheckErr(err, "Не могу подготовить запрос")
	defer stmt.Close()

	rows, err := stmt.Query("%"+confname+"%", "%"+projectname+"%", "%"+papername+"%", studentid)
	mlog.CheckErr(err, "Не могу выполнить запрос")
	defer rows.Close()

	switch format {
	case "HTML":
		ReportHTML(rows, w)
	case "XLSX":
		ReportXLSX(rows, w)
	case "DOCX":
		ReportDOCX(rows, w)
	default:
		ReportHTML(rows, w)
	}
}

func ReportHTML(rows *sql.Rows, w http.ResponseWriter) {
	sOut := ""
	nItogoConf, nItogoProject, nItogoPaper, nItogoStudent := 0, 0, 0, 0

	var fio sql.NullString
	var conference_point, project_point, paper_point sql.NullString

	for rows.Next() {
		err := rows.Scan(&fio, &conference_point, &project_point, &paper_point)
		mlog.CheckErr(err, "Не могу прочесть запись")
		nItogoStudent = 0

		if n, err := strconv.Atoi(conference_point.String); err == nil {
			nItogoConf += n
			nItogoStudent += n
		}
		if n, err := strconv.Atoi(project_point.String); err == nil {
			nItogoProject += n
			nItogoStudent += n
		}
		if n, err := strconv.Atoi(paper_point.String); err == nil {
			nItogoPaper += n
			nItogoStudent += n
		}

		sOut += fmt.Sprintf(`<tr>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%d</td>
			print`, fio.String, conference_point.String, project_point.String, paper_point.String, nItogoStudent)
	}

	sOut = `<table class="table table-striped">
		<thead>
			<tr><th>ФИО</th><th>Конференции</th><th>Проекты</th><th>Статьи</th><th>Все</th></tr>
		</thead>
		<tbody>` + sOut + fmt.Sprintf(`
			<tr><td><b>Итого:</b></td><td>%d</td><td>%d</td><td>%d</td><td>%d</td>
		</tbody>
	</table>`, nItogoConf, nItogoProject, nItogoPaper, nItogoConf+nItogoProject+nItogoPaper)
	fmt.Fprintf(w, "%v", sOut)
}

func ReportXLSX(rows *sql.Rows, w http.ResponseWriter) {
	nItogoConf, nItogoProject, nItogoPaper := 0, 0, 0
	f := excelize.NewFile()
	sheetName := "Отчет"
	
	index, err := f.NewSheet(sheetName)
	if err != nil {
		mlog.CheckErr(err, "Не могу создать лист")
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	headers := []string{"ФИО", "Конференции", "Проекты", "Статьи", "Все"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, h)
	}

	row := 2
	var fio sql.NullString
	var conference_point, project_point, paper_point sql.NullString

	for rows.Next() {
		err := rows.Scan(&fio, &conference_point, &project_point, &paper_point)
		mlog.CheckErr(err, "Не могу прочесть запись")

		confP, _ := strconv.Atoi(conference_point.String)
		projP, _ := strconv.Atoi(project_point.String)
		paperP, _ := strconv.Atoi(paper_point.String)
		total := confP + projP + paperP

		nItogoConf += confP
		nItogoProject += projP
		nItogoPaper += paperP

		cellA, _ := excelize.CoordinatesToCellName(1, row)
		cellB, _ := excelize.CoordinatesToCellName(2, row)
		cellC, _ := excelize.CoordinatesToCellName(3, row)
		cellD, _ := excelize.CoordinatesToCellName(4, row)
		cellE, _ := excelize.CoordinatesToCellName(5, row)

		f.SetCellValue(sheetName, cellA, fio.String)
		f.SetCellValue(sheetName, cellB, confP)
		f.SetCellValue(sheetName, cellC, projP)
		f.SetCellValue(sheetName, cellD, paperP)
		f.SetCellValue(sheetName, cellE, total)
		row++
	}

	cellA, _ := excelize.CoordinatesToCellName(1, row)
	cellB, _ := excelize.CoordinatesToCellName(2, row)
	cellC, _ := excelize.CoordinatesToCellName(3, row)
	cellD, _ := excelize.CoordinatesToCellName(4, row)
	cellE, _ := excelize.CoordinatesToCellName(5, row)

	f.SetCellValue(sheetName, cellA, "Итого:")
	f.SetCellValue(sheetName, cellB, nItogoConf)
	f.SetCellValue(sheetName, cellC, nItogoProject)
	f.SetCellValue(sheetName, cellD, nItogoPaper)
	f.SetCellValue(sheetName, cellE, nItogoConf+nItogoProject+nItogoPaper)

	sDate := time.Now().Format("2006-01-02")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment;filename="+sDate+"_Отчет.xlsx")
	
	if err := f.Write(w); err != nil {
		mlog.CheckErr(err, "Не могу записать Excel файл")
	}
}

func ReportDOCX(rows *sql.Rows, w http.ResponseWriter) {
	nItogoConf, nItogoProject, nItogoPaper := 0, 0, 0

	var fio sql.NullString
	var conference_point, project_point, paper_point sql.NullString

	for rows.Next() {
		err := rows.Scan(&fio, &conference_point, &project_point, &paper_point)
		mlog.CheckErr(err, "Не могу прочесть запись")

		if n, err := strconv.Atoi(conference_point.String); err == nil {
			nItogoConf += n
		}
		if n, err := strconv.Atoi(project_point.String); err == nil {
			nItogoProject += n
		}
		if n, err := strconv.Atoi(paper_point.String); err == nil {
			nItogoPaper += n
		}
	}

	rWord, err := docx.ReadDocxFile("mpage/report.docx")
	if err != nil {
		http.Error(w, "Файл шаблона report.docx не найден", http.StatusInternalServerError)
		return
	}
	defer rWord.Close()
	
	docx1 := rWord.Editable()
	docx1.Replace("old_1_1", strconv.Itoa(nItogoConf), -1)
	docx1.Replace("old_1_2", strconv.Itoa(nItogoProject), -1)
	docx1.Replace("old_1_3", strconv.Itoa(nItogoPaper), -1)

	sDate := time.Now().Format("2006-01-02")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment;filename="+sDate+"_Отчет.docx")
	docx1.Write(w)
}
