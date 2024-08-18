package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

const port = ":443"

var (
	certFile = filepath.Join("tmp", "certs", "tls.crt")
	keyFile  = filepath.Join("tmp", "certs", "tls.key")
)

func main() {
	// register the handler for /todo path
	http.HandleFunc("/todo", handleTodo)

	slog.Info("starting HTTP server", "port", port)
	err := http.ListenAndServeTLS(port, certFile, keyFile, nil)
	if err != nil {
		slog.Error("server failure", "error", err)
	}
}

// handleTodo handles the /todo URL path
func handleTodo(w http.ResponseWriter, _ *http.Request) {
	slog.Info("request received at path /todo")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	todoItems := getTodoList()
	err := json.NewEncoder(w).Encode(todoItems)
	if err != nil {
		slog.Error("failed to write response", "error", err)
	}
}

// getTodoList returns a list of todo items
func getTodoList() []todoItem {
	return []todoItem{
		{
			DueDate: time.Now().AddDate(0, 0, 7),
			ID:      uuid.NewString(),
			Title:   "write a todo-app",
		},
		{
			DueDate: time.Now().AddDate(0, 0, 8),
			ID:      uuid.NewString(),
			Title:   "define K8s manifests",
		},
		{
			DueDate: time.Now().AddDate(0, 0, 9),
			ID:      uuid.NewString(),
			Title:   "use certificates",
		},
	}
}

// todoItem defines a todo item
type todoItem struct {
	DueDate time.Time `json:"dueDate"`
	ID      string    `json:"id"`
	Title   string    `json:"title"`
}
