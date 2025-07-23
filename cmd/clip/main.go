package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("unix", "/tmp/clip.sock")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	_, err = io.Copy(conn, os.Stdin)

	if err != nil {
		fmt.Println("Error copy", err)
	}
}
