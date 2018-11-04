package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

type Payload struct {
	Data string
}

type User struct {
	Id    int
	Name  string
	Email string
}

type Config struct {
	Port     int
	Endpoint string
	Host     string
	User     string
	Password string
	Schema   string
}

var db *sql.DB

func init() {
	pwd, _ := os.Getwd()
	file, _ := os.Open(pwd + "/config.json")
	conf := Config{}
	err := json.NewDecoder(file).Decode(&conf)
	if err != nil {
		log.Fatal(err)
	}

	dbUrl := fmt.Sprintf("host=%s port=%d dbname=postgres user=%s password='%s' sslmode=disable search_path=%s",
		conf.Host, conf.Port, conf.User, conf.Password, conf.Schema)
	db, err = sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	http.HandleFunc("/api/users/", usersHandler)
	log.Println("listening")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		usersShow(w, r)
	case "DELETE":
		usersDelete(w, r)
	}
}

func usersDelete(w http.ResponseWriter, r *http.Request) {
	userId := strings.TrimPrefix(r.URL.Path, "/api/users/")
	query := fmt.Sprintf("SELECT * FROM test.user_del(%s)", userId)
	_, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(w, fmt.Sprintf(`{"id": %s}`, userId))
	log.Printf("Deleted user with id: %s", userId)
}

func usersShow(w http.ResponseWriter, r *http.Request) {
	userId := strings.TrimPrefix(r.URL.Path, "/api/users/")
	query := fmt.Sprintf("SELECT user_get  FROM test.user_get(%s)", userId)
	payload := Payload{}
	err := db.QueryRow(query).Scan(&payload.Data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		log.Println(err)
	}
	user := User{}
	json.Unmarshal([]byte(payload.Data), &user)

	json.NewEncoder(w).Encode(&user)
	log.Printf("Found user with id: %s", userId)
}
