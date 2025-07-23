package server

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ConnInfo struct {
	ID   string
	Conn *websocket.Conn
}

type Hub struct {
	connections map[string]ConnInfo
	mu          sync.RWMutex
}

func NewHub() Hub {
	return Hub{connections: make(map[string]ConnInfo)}
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

type ListenMsgHandler = func(int, []byte)

func (h *Hub) Listen(c ConnInfo, handler ListenMsgHandler) {
	for {
		mt, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s from ID=%s", message, c.ID)
		handler(mt, message)
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

type HandleFunc func(w http.ResponseWriter, r *http.Request)
