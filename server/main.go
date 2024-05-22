package main

import (
	"log"
	"net/http"
	"time"
	"sync"
	"github.com/gorilla/websocket"
	"encoding/json"
)

type Message struct {
	User string 		`json:"user"`
	Timestamp time.Time `json:"timestamp"`
}

type Notification struct {
	MessageType string    `json:"type"`
	User        string    `json:"user"`
}

var upgrader = websocket.Upgrader{}
var mutex sync.Mutex
var MessageInfo []Message
var connections []*websocket.Conn

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }
	defer conn.Close()

	connections = append(connections, conn)

	for {
		_, mes, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
		}

		var message Message
		if err := json.Unmarshal(mes, &message); err != nil {
			log.Println(err)
			continue
		}

		mutex.Lock()
		MessageInfo = append(MessageInfo, message)
		mutex.Unlock()

		for _, c := range connections { 
			err := c.WriteJSON(message)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func main() {
	http.HandleFunc("/", handleConnections)
	err := http.ListenAndServe("localhost:3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
