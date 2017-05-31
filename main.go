package main

import (
	"net/http"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"html/template"
)

var db2 *sql.DB
var tmpl *template.Template

func init() {
	var err error
	db2, err = sql.Open("mysql", "myrss@/myrss?charset=utf8")
	if err != nil {
		log.Fatal(err)
	}

	tmpl, err = template.ParseFiles("index.gohtml")
}

func rootHandler(w http.ResponseWriter, r *http.Request) {

	rows, err := db2.Query("SELECT * FROM `articles` ORDER BY `created` DESC limit 100")
	defer rows.Close()
	if err != nil {
		log.Fatalln(err)
	}

	cols, _ := rows.Columns()
	pointers := make([]interface{}, len(cols))
	container := make([]interface{}, len(cols))

	for i := range pointers {
		pointers[i] = &container[i]
	}

	for rows.Next() {

		rows.Scan(pointers ...)
	}

	tmpl.Execute(w, container)
	//fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func main() {
	defer db2.Close()
	//fmt.Println("Hello World")

	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":8080", nil)
}
