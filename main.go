package main

import (
	"log"
	"net/http"

	"github.com/atrevbot/prl/app"
	"github.com/boltdb/bolt"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

const DB_NAME = "data.db"

func main() {
	// Load env file
	env, err := godotenv.Read()
	if err != nil {
		panic(err)
	}

	// Open DB and create required repositories
	db, err := bolt.Open(DB_NAME, 0600, nil)
	if err != nil {
		panic(err)
	}
	symptomRepo, err := app.NewSymptomRepo(db)
	if err != nil {
		panic(err)
	}
	eventRepo, err := app.NewEventRepo(db)
	if err != nil {
		panic(err)
	}

	// Create server and attach routes
	s := app.Server{symptomRepo, eventRepo, mux.NewRouter(), env}
	s.Routes()

	// Start server
	secret := []byte(env["SECRET_KEY"])
	secure := csrf.Secure(env["ENVIRONMENT"] != "development")
	log.Fatal(http.ListenAndServe(":8080", csrf.Protect(secret, secure)(s.Router)))
}
