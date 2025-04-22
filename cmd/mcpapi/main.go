package main

import (
	"fmt"
	"net/http"
	"os"
	"github.com/go-chi/chi/v5"
	"github.com/kwakuoseikwakye/go-mcps/internal/mcp"
	"github.com/kwakuoseikwakye/go-mcps/internal/slack"
	"github.com/kwakuoseikwakye/go-mcps/internal/github"
)

func main() {
	// Register servers
	mcp.RegisterServer(slack.New())
	mcp.RegisterServer(github.New())

	r := chi.NewRouter()
	r.Route("/api/v1/{server}", func(r chi.Router) {
		r.Get("/contexts", handleListContexts)
		r.Post("/send", handleSendMessage)
		r.Get("/receive", handleReceiveMessage)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Starting MCP REST API server on port", port)
	http.ListenAndServe(":"+port, r)
}

func handleListContexts(w http.ResponseWriter, r *http.Request) {
	serverName := chi.URLParam(r, "server")
	srv, ok := mcp.GetServer(serverName)
	if !ok {
		http.Error(w, "Unknown server: "+serverName, 400)
		return
	}
	_ = srv.Connect(make(map[string]string))
	contexts, err := srv.ListContexts()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	for _, ctx := range contexts {
		_, _ = w.Write([]byte(ctx + "\n"))
	}
}

func handleSendMessage(w http.ResponseWriter, r *http.Request) {
	serverName := chi.URLParam(r, "server")
	srv, ok := mcp.GetServer(serverName)
	if !ok {
		http.Error(w, "Unknown server: "+serverName, 400)
		return
	}
	_ = srv.Connect(make(map[string]string))
	ctx := r.URL.Query().Get("context")
	msg := r.URL.Query().Get("message")
	if ctx == "" || msg == "" {
		http.Error(w, "Missing context or message", 400)
		return
	}
	if err := srv.SendMessage(ctx, msg); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Write([]byte("OK"))
}

func handleReceiveMessage(w http.ResponseWriter, r *http.Request) {
	serverName := chi.URLParam(r, "server")
	srv, ok := mcp.GetServer(serverName)
	if !ok {
		http.Error(w, "Unknown server: "+serverName, 400)
		return
	}
	_ = srv.Connect(make(map[string]string))
	ctx := r.URL.Query().Get("context")
	if ctx == "" {
		http.Error(w, "Missing context", 400)
		return
	}
	ch, err := srv.ReceiveMessage(ctx)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	for msg := range ch {
		_, _ = w.Write([]byte(fmt.Sprintf("[%s] %s: %s\n", msg.Time, msg.User, msg.Text)))
	}
}