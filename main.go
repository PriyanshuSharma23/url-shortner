package main

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"unsafe"

	"github.com/go-chi/chi/v5"
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

var alphabet = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generate(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	for i := 0; i < size; i++ {
		b[i] = alphabet[b[i]%byte(len(alphabet))]
	}
	return *(*string)(unsafe.Pointer(&b))
}

func main() {
	log := log.New(os.Stdout, "[url-shortener] ", log.LstdFlags|log.Lshortfile)
	dbDsn := os.Getenv("DB_DSN")

	db, err := sql.Open("postgres", dbDsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	defer db.Close()

	log.Println("Database connection established.")

	mux := chi.NewMux()

	mux.Get("/{shortUrl}", func(w http.ResponseWriter, r *http.Request) {
		shortUrl := chi.URLParam(r, "shortUrl")

		url := new(string)

		sqlCmd := "SELECT url FROM urls WHERE short_url = $1"
		err := db.QueryRow(sqlCmd, shortUrl).Scan(url)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}

			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Location", *url)
		w.WriteHeader(http.StatusFound)
	})

	mux.Post("/", func(w http.ResponseWriter, r *http.Request) {
		addURL := r.URL.Query().Get("url")

		parsedURL, err := url.Parse(addURL)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if parsedURL.Scheme == "" || parsedURL.Host == "" {
			log.Println("Invalid URL: " + addURL)
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}

		shortUrl := generate(16)
		sql := "INSERT INTO urls (url, short_url) VALUES ($1, $2)"
		_, err = db.Exec(sql, parsedURL.String(), shortUrl)

		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortUrl))
	})

	log.Println("Server started.")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
