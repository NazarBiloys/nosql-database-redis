package main

import (
	"fmt"
	"github.com/NazarBiloys/nosql-database-redis/internal/service"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := service.MakeUser(); err != nil {
			panic(err)
		}
	}

	if r.Method == "GET" {
		users, err := service.FetchFirstFiveUser(5)
		if err != nil {
			panic(err)
		}

		_, err = fmt.Fprintf(w, "Users : %s", users)
		if err != nil {
			return
		}
	}
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":90", nil))
}
