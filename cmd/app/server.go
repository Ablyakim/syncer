package main

import (
	"context"
	"log"
	"net/http"

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

type ConnInfo struct {
	ID   string
	Conn *websocket.Conn
}

var connections = map[string]ConnInfo{}

type Hub struct {
	connections map[string]ConnInfo
	// Add mutex
}

func (h *Hub) RegisterConnection(c ConnInfo) {
	if _, ok := h.connections[c.ID]; !ok {
		h.connections[c.ID] = c
	}
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

func (h *Hub) UnRegisterConnection(c ConnInfo) {
	panic("Implement me")
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
