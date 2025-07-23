package serverhandler

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"

	"github.com/Ablyakim/syncer/internal/server"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// TODO hack
const watcher = "/etc/profiles/per-user/ab/bin/theme-watcher.swift"

func ThemeHandler(w http.ResponseWriter, r *http.Request) {
	hub := server.NewHub()
	c, err := server.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	connID := uuid.NewString()
	ci := server.ConnInfo{ID: connID, Conn: c}
	hub.RegisterConnection(ci)
	defer hub.RemoveConnection(ci)

	go func() {
		currTheme, err := getCurrentTheme(context.Background())
		if err != nil {
			currTheme = []byte("Dark")
			log.Println(err)
		}
		err = hub.WriteMessage(connID, websocket.TextMessage, currTheme)
		if err != nil {
			log.Println(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go watchTheme(ctx, &hub)

	hub.Listen(ci, func(mt int, message []byte) {
		hub.Broadcast(ci.ID, mt, message)
	})
}

func watchTheme(ctx context.Context, hub *server.Hub) {
	cmd := exec.CommandContext(ctx, watcher)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println(err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Println(err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		hub.Broadcast("", websocket.TextMessage, scanner.Bytes())
		fmt.Println("Theme switched:", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
		return
	}

	if err := cmd.Wait(); err != nil {
		log.Println(err)
		return
	}
}

func getCurrentTheme(ctx context.Context) ([]byte, error) {
	cmd := exec.CommandContext(ctx, watcher, "get")

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
