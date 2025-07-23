package main

import (
	"log"
	"net/http"

	"github.com/Ablyakim/syncer/internal/serverhandler"
)

type Config struct {
	Addr string
	Path string
}

func server(addr string) {
	handlersMap := make(map[string]serverhandler.HandleFunc)
	handlersMap["theme"] = serverhandler.ThemeHandler
	handlersMap["clip"] = serverhandler.ClipHandler

	for path, h := range handlersMap {
		http.HandleFunc("/"+path, h)
	}

	log.Fatal(http.ListenAndServe(addr, nil))
}
