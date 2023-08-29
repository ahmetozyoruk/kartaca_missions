package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/ahmetozyoruk/myapp/config"
	"github.com/ahmetozyoruk/myapp/sftp"
	_ "github.com/go-sql-driver/mysql"
)


type myHandler struct{}

type MovieTable struct {
	id       int    `json:"id"`
	title    string `json:"title"`
	synopsis string `json:"synopsis"`
	genre    string `json:"genre"`
	year     int    `json:"year"`
	duration int    `json:"duration"`
}

func (*myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "URL: "+r.URL.String())
}

func CheckDB(w http.ResponseWriter, r *http.Request) {

	config := config.InitConfig()
	
	// "root:ahmet@tcp(mysql:3306)/myapp"

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
		io.WriteString(w, "OK")
	}else{
		io.WriteString(w, "NOK")
	}
}

func CheckSite(w http.ResponseWriter, r *http.Request) {

	config := config.InitConfig()

	resp, err := http.Get(config.Website.Url)
	// resp, err := http.NewRequest("GET", config.Website.Url, nil)

	if err != nil {
		log.Printf("Body read error, %v", err)
		io.WriteString(w, "NOK")
		// w.WriteHeader(500) // Return 500 Internal Server Error.
		return
	} 

	defer resp.Body.Close()

	// fmt.Println("The status code we got is:", resp.Response.StatusCode)

	fmt.Println("The status code we got is:", resp.StatusCode)

	io.WriteString(w, "OK")
}

func CheckServer(w http.ResponseWriter, r *http.Request) {

	config := config.InitConfig()


	initValues := sftp.Config{
		Username:     config.Server.Username,
		Password:     config.Server.Password, // required only if password authentication is to be used
		Server:       config.Server.Host+":"+config.Server.Port,
		KeyExchanges: []string{"diffie-hellman-group-exchange-sha256", "diffie-hellman-group14-sha256"}, // optional
		Timeout:      time.Second * 30,                                                                  // 0 for not timeout
	}

	client, err := sftp.New(initValues)
	if err != nil {
		log.Printf("Body read error, %v", err)
		io.WriteString(w, "NOK")
		return
	}
	defer client.Close()

	// Download remote file.
	file, err := client.Download("/home/kubernetmachinetwo/asd.txt")
	if err != nil {
		// log.Fatalln(err)
		log.Printf("Body read error, %v", err)
		io.WriteString(w, "NOK")
		return
	}
	defer file.Close()

	fmt.Println("there is a file: /home/kubernetmachinetwo/asd.txt")
	io.WriteString(w, "OK")
}


func main() {
	
	mux := http.NewServeMux()

	// Register routes and register handlers in this form.
	mux.Handle("/", &myHandler{})

	mux.HandleFunc("/checkDB", CheckDB)
	mux.HandleFunc("/checkSite", CheckSite)
	mux.HandleFunc("/checkServer", CheckServer)

	fmt.Println("Server is listening at 8080 port")
	//http.ListenAndServe uses the default server structure.
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
