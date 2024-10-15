package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	port := os.Getenv("PORT")
	r := mux.NewRouter()
	r.HandleFunc("/", getBlock).Methods("GET")
	r.HandleFunc("/", writeBlock).Methods("POST")
	r.HandleFunc("/new", newBlock).Methods("GET")
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err, "error when listening and serve the server")
	}
	log.Println("listing and serve on ", port)
}
