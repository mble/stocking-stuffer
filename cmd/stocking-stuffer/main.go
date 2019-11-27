package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	_ "github.com/lib/pq"
	"github.com/mble/stocking-stuffer/internal/bells"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

type Credentials struct {
	Password string `json:"password" db:"password"`
	Username string `json:"username" db:"username"`
}

func main() {
	initDB()
	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(bells.LogrusLogger())
	r.Use(bells.UserID)
	r.Use(bells.RateLimitIP)
	r.Use(bells.RateLimitDevice)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(15 * time.Second))

	r.Post("/signin", Signin)
	r.Post("/signup", Signup)

	s := &http.Server{
		Addr:         ":8000",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	log.Fatal(s.ListenAndServe())
}

func initDB() {
	var err error
	db, err = sql.Open("postgres", "dbname=stocking-stuffer_dev sslmode=disable")
	err = db.Ping()
	if err != nil {
		panic(err)
	}
}

func Signin(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		bells.LogEntrySetField(r, "error", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	result := db.QueryRow("select password from users where username=$1", creds.Username)
	if err != nil {
		bells.LogEntrySetField(r, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	storedCreds := &Credentials{}
	err = result.Scan(&storedCreds.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			bells.LogEntrySetField(r, "failure_reason", "user_not_found")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hash := fmt.Sprintf("%x", md5.Sum([]byte(creds.Password)))
	result = db.QueryRow("select 1 from pwned_passwords where password_md5_hash = $1", hash)
	var ok interface{}
	err = result.Scan(&ok)
	if ok != nil {
		bells.LogEntrySetField(r, "failure_reason", "leaked_password")
		// To avoid leaking information, lets do some useless work.
		bcrypt.GenerateFromPassword([]byte("much-wow-such-work"), 8)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(storedCreds.Password), []byte(creds.Password)); err != nil {
		bells.LogEntrySetField(r, "failure_reason", "bad_password")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)

	if _, err = db.Query("insert into users (username, password) values ($1, $2)", creds.Username, string(hashedPassword)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
