package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	MessageType string    `json:"messageType"`
	User        string    `json:"user"`
	Timestamp   time.Time `json:"timestamp"`
}
type Notification struct {
	MessageType string `json:"type"`
	User        string `json:"user"`
}

var upgrader = websocket.Upgrader{}
var mutex sync.Mutex
var MessageInfo []Message
var connections = make(map[*websocket.Conn]string) // map to store connections and their associated users

func notifyAll(notification Notification) {
	mutex.Lock()
	defer mutex.Unlock()

	for c := range connections {
		if err := c.WriteJSON(notification); err != nil {
			log.Println(err)
		}
	}
}

func handleDisconnection(conn *websocket.Conn) {
	mutex.Lock()
	defer mutex.Unlock()
	
	user := connections[conn]
	delete(connections, conn) // remove the connection from the map

	notification := Notification{MessageType: "User Disconnected", User: user}
	notifyAll(notification)
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	
	defer func() {
		handleDisconnection(conn)
	}()

	for {
		_, mes, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			handleDisconnection(conn)
			break
		}

		var message Message
		if err := json.Unmarshal(mes, &message); err != nil {
			log.Println(err)
			continue
		}

		mutex.Lock()
		if message.MessageType == "New User Joined" {
			connections[conn] = message.User
			notification := Notification{MessageType: "New User Joined", User: message.User}
			notifyAll(notification)
		}

		MessageInfo = append(MessageInfo, message)
		mutex.Unlock()

		for c := range connections {
			err := c.WriteJSON(message)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func handleNextRound(w http.ResponseWriter, r *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()

	MessageInfo = make([]Message, 0)
}

func main() {
	http.HandleFunc("/", handleConnections)
	http.HandleFunc("/next-round", handleNextRound)
	
	err := http.ListenAndServe("localhost:3000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
