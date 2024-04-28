package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/joeychilson/ai-chat/ui"
	"github.com/joeychilson/ai/anthropic"
)

type Server struct {
	anthropic *anthropic.Client
}

func New(anthropic *anthropic.Client) *Server {
	return &Server{anthropic: anthropic}
}

func (s *Server) Router() http.Handler {
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

	router.Mount("/", ui.Handler())
	router.Post("/chat", s.handleChat())

	return router
}

func (s *Server) handleChat() http.HandlerFunc {
	type request struct {
		Message string `json:"message"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
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

		err := s.anthropic.ChatStream(r.Context(), llmReq, func(ctx context.Context, event anthropic.Event) {
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
