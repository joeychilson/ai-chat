package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
				e := &Event{
					Event: []byte("error"),
					Data:  []byte("{\"message\": \"Internal Server Error\"}"),
				}
				if err := e.Write(w); err != nil {
					log.Println("Unable to write event:", err)
					return
				}
				w.(http.Flusher).Flush()
				return
			}
			e := &Event{Data: jsonStr}
			if err := e.Write(w); err != nil {
				log.Println("Unable to write event:", err)
				return
			}
			w.(http.Flusher).Flush()
		})

		if err != nil {
			log.Println("Unable to chat:", err)
			e := &Event{
				Event: []byte("error"),
				Data:  []byte("{\"message\": \"Internal Server Error\"}"),
			}
			if err := e.Write(w); err != nil {
				log.Println("Unable to write event:", err)
				return
			}
			w.(http.Flusher).Flush()
			return
		}
	}
}

// Event represents a server-sent event.
type Event struct {
	Event []byte
	Data  []byte
}

// WriteTo writes the event to the writer.
func (e *Event) Write(w io.Writer) error {
	if len(e.Data) == 0 {
		return nil
	}
	if len(e.Data) > 0 {
		sd := bytes.Split(e.Data, []byte("\n"))
		for i := range sd {
			if _, err := fmt.Fprintf(w, "data: %s\n", sd[i]); err != nil {
				return err
			}
		}
		if len(e.Event) > 0 {
			if _, err := fmt.Fprintf(w, "event: %s\n", e.Event); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprint(w, "\n"); err != nil {
		return err
	}
	return nil
}
