package main

import (
	"log"
	"net/http"
)

func main() {
	r := NewRouter()

	log.Println("Server is running on port 8000")
	
	if err := http.ListenAndServe(":8000", r); err != nil {
		log.Fatal("failed to start server")
	}
}