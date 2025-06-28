package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed dist/*
var content embed.FS

var clients = make(map[*websocket.Conn]bool)
var clientsMu sync.Mutex

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // Allow any origin
}

func BroadcastLikeCount(count uint64) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for conn := range clients {
		err := conn.WriteJSON(map[string]any{
			"type":  "like_update",
			"likes": count,
		})
		if err != nil {
			conn.Close()
			delete(clients, conn)
		}
	}
}

func New() *echo.Echo {
	e := echo.New()
	e.Use(middleware.Logger())

	// Subset embedded filesystem
	staticFS, err := fs.Sub(content, "dist")
	if err != nil {
		e.Logger.Fatal(err)
	}
	fsys := http.FS(staticFS)

	// WebSocket endpoint
	e.GET("/ws", func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		clientsMu.Lock()
		clients[conn] = true
		clientsMu.Unlock()
		return nil
	})

	// Static file and SPA fallback handler
	e.GET("/*", echo.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		f, err := staticFS.Open(path)
		if err == nil {
			f.Close()
			http.FileServer(fsys).ServeHTTP(w, r)
			return
		}
		r.URL.Path = "/index.html"
		http.FileServer(fsys).ServeHTTP(w, r)
	})))

	return e
}
