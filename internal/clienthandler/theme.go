package clienthandler

import (
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

type ClientFunc func(*websocket.Conn, *http.Response) <-chan struct{}

func ThemeClient(c *websocket.Conn, _ *http.Response) <-chan struct{} {
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				errorLog.Println(err)
				return
			}
			_, err = os.Stdout.Write(message)
			if err != nil {
				errorLog.Println(err)
				return
			}
		}
	}()

	return done
}
