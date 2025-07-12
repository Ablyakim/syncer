package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/gorilla/websocket"
)

// TODO hack
const watcher = "/etc/profiles/per-user/ab/bin/theme-watcher.swift"

func watchTheme(ctx context.Context) {
	cmd := exec.CommandContext(ctx, watcher)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		hub.Broadcast("", websocket.TextMessage, scanner.Bytes())
		fmt.Println("Theme switched:", scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
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
