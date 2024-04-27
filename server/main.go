package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"

	"github.com/joeychilson/ai/anthropic"
)

func main() {
	anthropicClient := anthropic.New(os.Getenv("ANTHROPIC_API_KEY"))

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"POST"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
	}))

	router.Post("/chat", handleChat(anthropicClient))

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

type ChatRequest struct {
	Message string `json:"message"`
}

func handleChat(client *anthropic.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Println("Unable to parse request body:", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		llmReq := &anthropic.ChatRequest{
			Model: anthropic.ModelClaude3_Haiku,
			Messages: []anthropic.Message{
				anthropic.UserMessage{Content: []anthropic.Content{anthropic.TextContent{Text: req.Message}}},
			},
			MaxTokens: 1054,
		}

		err := client.ChatStream(r.Context(), llmReq, func(ctx context.Context, event anthropic.Event) {
			jsonStr, err := json.Marshal(event)
			if err != nil {
				log.Println("Unable to marshal event:", err)
				fmt.Fprintf(w, "event: error\ndata: {\"message\": \"Internal Server Error\"}\n\n")
				w.(http.Flusher).Flush()
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", string(jsonStr))
			w.(http.Flusher).Flush()
		})

		if err != nil {
			log.Println("Unable to chat:", err)
			fmt.Fprintf(w, "event: error\ndata: {\"message\": \"Internal Server Error\"}\n\n")
			w.(http.Flusher).Flush()
			return
		}
	}
}
