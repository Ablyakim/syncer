package serverhandler

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/Ablyakim/syncer/internal/server"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.design/x/clipboard"
)

var (
	infoLog  = log.New(os.Stdout, "INFO: ", log.LstdFlags)
	errorLog = log.New(os.Stderr, "ERROR: ", log.LstdFlags)
)

type HandleFunc func(w http.ResponseWriter, r *http.Request)

func ClipHandler(w http.ResponseWriter, r *http.Request) {
	hub := server.NewHub()
	c, err := server.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		errorLog.Print(err)
		return
	}
	defer c.Close()

	connID := uuid.NewString()
	ci := server.ConnInfo{ID: connID, Conn: c}
	hub.RegisterConnection(ci)
	defer hub.RemoveConnection(ci)
	if err = clipboard.Init(); err != nil {
		errorLog.Println(err)
		return
	}

	mu := sync.Mutex{}
	var currData []byte

	go func() {
		ch := clipboard.Watch(context.Background(), clipboard.FmtText)
		for data := range ch {
			mu.Lock()
			if !bytes.Equal(currData, data) {
				infoLog.Printf("Read from clipboard bytes=%d\n", len(data))
				currData = data
				err = hub.WriteMessage(connID, websocket.TextMessage, data)
				if err != nil {
					errorLog.Println(err)
				}
			} else {
				infoLog.Println("Clipboard synced")
			}
			mu.Unlock()
		}
	}()

	hub.Listen(ci, func(_ int, message []byte) {
		mu.Lock()
		infoLog.Printf("Received clip from server bytes=%d\n", len(message))
		currData = message
		clipboard.Write(clipboard.FmtText, currData)
		mu.Unlock()
	})
}
