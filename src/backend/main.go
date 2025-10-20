package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	r := NewRouter()
	port := os.Getenv("PORT")
	
	if port == "" {
		port = "8000"
	}

	log.Printf("Server is running on port %s", port)
	
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), r); err != nil {
		log.Fatal("failed to start server")
	}
}