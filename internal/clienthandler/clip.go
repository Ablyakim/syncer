package clienthandler

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"

	"github.com/gorilla/websocket"
)

const localSockPath = "/tmp/clip.sock"

var (
	infoLog  = log.New(os.Stdout, "INFO: ", log.LstdFlags)
	errorLog = log.New(os.Stderr, "ERROR: ", log.LstdFlags)
)

func ClipClient(wsCon *websocket.Conn, _ *http.Response) <-chan struct{} {
	done := make(chan struct{})
	sendToServer := serverSender(wsCon)

	go func() {
		defer close(done)
		for {
			_, message, err := wsCon.ReadMessage()
			if err != nil {
				errorLog.Println("Error", err)
				return
			}
			infoLog.Printf("Retrieved clip from mac %s\n", message)

			sendToTmux(message)
		}
	}()

	go func() {
		defer close(done)
		_ = os.Remove(localSockPath)
		l, err := net.Listen("unix", localSockPath)
		if err != nil {
			panic(err)
		}
		defer l.Close()
		defer os.Remove(localSockPath)

		infoLog.Println("clipd running...")

		for {
			conn, err := l.Accept()
			if err != nil {
				errorLog.Println("Error accepting connection", err)
				continue
			}

			go func(c net.Conn) {
				defer c.Close()
				scanner := bufio.NewScanner(c)
				for scanner.Scan() {
					msg := scanner.Bytes()
					infoLog.Printf("Message from xclip bytes=%d\n", len(msg))
					go sendToServer(msg)
					go sendToTmux(msg)
				}
			}(conn)
		}
	}()

	return done
}

func serverSender(wsCon *websocket.Conn) func(msg []byte) {
	return func(msg []byte) {
		infoLog.Printf("Sending to server bytes=%d\n", len(msg))
		err := wsCon.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			errorLog.Printf("Failed send to server Error: %v\n", err)
		}
	}
}

func sendToTmux(msg []byte) {
	infoLog.Printf("Sending to tmux bytes=%d\n", len(msg))

	cmd := exec.Command("tmux", "load-buffer", "-")
	cmd.Stdin = bytes.NewReader(msg)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		errorLog.Printf("tmux stdout: %s\n", stdout.String())
		errorLog.Printf("tmux stderr: %s\n", stderr.String())
		errorLog.Printf("tmux command failed Error: %v\n", err)
	}
}
