package main

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var hub Hub

type Config struct {
	Addr string
	Path string
}

type ConnInfo struct {
	ID   string
	Conn *websocket.Conn
}

type Hub struct {
	connections map[string]ConnInfo
	mu          sync.RWMutex
}

func (h *Hub) RegisterConnection(c ConnInfo) {
	if _, ok := h.connections[c.ID]; !ok {
		h.mu.Lock()
		h.connections[c.ID] = c
		h.mu.Unlock()
	}
}

func (h *Hub) RemoveConnection(c ConnInfo) {
	log.Printf("Remove connection %s from hub", c.ID)
	h.mu.Lock()
	delete(h.connections, c.ID)
	h.mu.Unlock()
}

func (h *Hub) Listen(c ConnInfo) {
	for {
		mt, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s from ID=%s", message, c.ID)
		h.Broadcast(c.ID, mt, message)
	}
}

func (h *Hub) WriteMessage(connID string, mt int, msg []byte) error {
	if conn, ok := h.connections[connID]; ok {
		return conn.Conn.WriteMessage(mt, msg)
	}

	return nil
}

func (h *Hub) Broadcast(fromConnID string, mt int, msg []byte) {
	for connId, c := range h.connections {
		if fromConnID == connId {
			continue
		}

		err := h.WriteMessage(c.ID, mt, msg)
		if err != nil {
			log.Printf("Can't send msg to client %s, error %v", c.ID, err)
		}
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	currTheme, err := getCurrentTheme(context.Background())
	if err != nil {
		currTheme = []byte("Dark")
		log.Println(err)
	}

	connID := uuid.NewString()
	ci := ConnInfo{ID: connID, Conn: c}
	hub.RegisterConnection(ci)
	defer hub.RemoveConnection(ci)

	err = hub.WriteMessage(connID, websocket.TextMessage, currTheme)
	if err != nil {
		log.Println(err)
	}

	hub.Listen(ci)
}

func server(addr, path string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hub = Hub{connections: make(map[string]ConnInfo)}
	http.HandleFunc("/"+path, handler)

	go watchTheme(ctx)
	log.Fatal(http.ListenAndServe(addr, nil))
}
