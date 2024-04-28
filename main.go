package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joeychilson/ai-chat/server"
	"github.com/joeychilson/ai/anthropic"
)

func main() {
	anthropicClient := anthropic.New(os.Getenv("ANTHROPIC_API_KEY"))
	server := server.New(anthropicClient)

	log.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", server.Router()); err != nil {
		log.Fatal(err)
	}
}
