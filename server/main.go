package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w,r, nil)
		if err != nil {
			log.Println(err)
		}

		defer conn.Close()

	})

	http.ListenAndServe("localhost:3000", nil)
}
