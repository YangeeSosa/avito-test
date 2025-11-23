package app

import (
	"net/http"

	"github.com/test-avito/internal/db"
	transport "github.com/test-avito/internal/http"
	"github.com/test-avito/internal/service"
)

type App struct {
	httpHandler http.Handler
}

func New() *App {
	store := db.NewMemoryStore()
	svc := service.New(store)
	server := transport.NewServer(svc)

	return &App{
		httpHandler: server.Router(),
	}
}

func (a *App) Handler() http.Handler {
	return a.httpHandler
}
