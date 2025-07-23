package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/Ablyakim/syncer/internal/clienthandler"
	"github.com/gorilla/websocket"
)

func client(addr, path string) {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	clients := make(map[string]clienthandler.ClientFunc)
	clients["theme"] = clienthandler.ThemeClient
	clients["clip"] = clienthandler.ClipClient
	client, ok := clients[path]
	if !ok {
		log.Fatalf("Client for path %v not found", path)
	}

	u := url.URL{Scheme: "ws", Host: addr, Path: "/" + path}
	log.Printf("connecting to %s", u.String())

	c, r, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	done := client(c, r)

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
