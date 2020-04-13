package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"github.com/lib/pq"

	"github.com/subosito/gotenv"

	"github.com/gorilla/mux"
)

type Book struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Year   string `json:"year"`
}

var books []Book
var db *sql.DB

func init() {
	gotenv.Load()
}
func main() {
	pgURL, err := pq.ParseURL(os.Getenv("PG_URL"))
	logFatal(err)
	log.Println(pgURL)

	db, err = sql.Open("postgres", pgURL)
	logFatal(err)

	err = db.Ping()
	logFatal(err)
	log.Println(db)

	router := mux.NewRouter()
	books = append(books, Book{ID: 1, Title: "Title 1", Author: "Author1", Year: "Year1"},
		Book{ID: 2, Title: "Title 2", Author: "Author 2", Year: "Year 2"},
		Book{ID: 3, Title: "Title 3", Author: "Author 3", Year: "Year 3"},
		Book{ID: 4, Title: "Title 4", Author: "Author 4", Year: "Year 4"})

	router.HandleFunc("/books", getBooks).Methods("GET")
	router.HandleFunc("/books/{id}", getBook).Methods("GET")
	router.HandleFunc("/books", updateBook).Methods("PUT")
	router.HandleFunc("/books", addBook).Methods("POST")
	router.HandleFunc("/books/{id}", removeBook).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":8000", router))
}

// func getBooks(w http.ResponseWriter, r *http.Request) {
// json.NewEncoder(w).Encode(books)
// }

func getBooks(w http.ResponseWriter, r *http.Request) {
	var book Book
	books := []Book{}
	rows, err := db.Query("select * from books")
	logFatal(err)
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year)
		logFatal(err)
		books = append(books, book)
	}
	json.NewEncoder(w).Encode(books)

}

// func getBook(w http.ResponseWriter, r *http.Request) {
// 	params := mux.Vars(r)
// 	id, _ := strconv.Atoi(params["id"])
// 	for _, book := range books {
// 		if book.ID == id {
// 			json.NewEncoder(w).Encode(&book)
// 		}
// 	}
// }

func getBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	params := mux.Vars(r)
	rows := db.QueryRow("select * from books where id = $1", params["id"])
	err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year)
	logFatal(err)
	json.NewEncoder(w).Encode(&book)
}

// func addBook(w http.ResponseWriter, r *http.Request) {
// 	var book Book
// 	json.NewDecoder(r.Body).Decode(&book)
// 	books = append(books, book)
// 	json.NewEncoder(w).Encode(books)
// }

func addBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	var bookID int
	json.NewDecoder(r.Body).Decode(&book)
	err := db.QueryRow("insert into books (title, author, year) values ($1, $2, $3) returning id;", book.Title, book.Author, book.Year).Scan(&bookID)
	logFatal(err)
	json.NewEncoder(w).Encode(bookID)
}

// func updateBook(w http.ResponseWriter, r *http.Request) {
// 	var book Book
// 	json.NewDecoder(r.Body).Decode(&book)
// 	for i, item := range books {
// 		if item.ID == book.ID {
// 			books[i] = book
// 		}
// 	}
// 	json.NewEncoder(w).Encode(books)
// }

func updateBook(w http.ResponseWriter, r *http.Request) {
	var book Book

	json.NewDecoder(r.Body).Decode(&book)
	db.Exec("update books set title=$1, author=$2, year=$3 where id= $4 returning id", &book)
	for i, item := range books {
		if item.ID == book.ID {
			books[i] = book
		}
	}
	json.NewEncoder(w).Encode(books)
}
func removeBook(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])
	log.Println(reflect.TypeOf(id))
	for i, item := range books {
		if item.ID == id {
			books = append(books[:i], books[i+1:]...)
		}
	}
	json.NewEncoder(w).Encode(books)
}

func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
