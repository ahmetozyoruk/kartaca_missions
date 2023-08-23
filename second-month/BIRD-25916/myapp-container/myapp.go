package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/ahmetozyoruk/myapp/config"
	_ "github.com/go-sql-driver/mysql"
)

type myHandler struct{}

func (*myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "URL: "+r.URL.String())
}

func CheckDB(w http.ResponseWriter, r *http.Request) {

	// "root:ahmet@tcp(mysql:3306)/myapp"
	config := config.InitConfig()

	db, err := sql.Open("mysql", config.Database.DBUsername+":"+config.Database.DBPassword+"@tcp("+config.Database.DBHost+":"+config.Database.DBPort+")/"+config.Database.DBName)
	if err != nil {
		// panic(err.Error())
		log.Printf("Body read error, %v", err)
		io.WriteString(w, "NOK")
		return
	}

	defer db.Close()
	num_row, err := db.Query("SELECT  COUNT(*) FROM movies")

	if err != nil {
		log.Printf("Body read error, %v", err)
		io.WriteString(w, "NOK")
		return
	}
	defer num_row.Close()

	var count int

	for num_row.Next() {
		if err := num_row.Scan(&count); err != nil {
			log.Fatal(err)
			return
		}
	}

	if count > 10 {
		io.WriteString(w, "NOK")
	} else {
		io.WriteString(w, "OK")
	}
}

func main() {

	mux := http.NewServeMux()

	// Register routes and register handlers in this form.
	mux.Handle("/", &myHandler{})

	mux.HandleFunc("/checkDB", CheckDB)

	fmt.Println("Server is listening at 8080 port")
	//http.ListenAndServe uses the default server structure.
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
