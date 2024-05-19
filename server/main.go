package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// User represents a connected user
type User struct {
	Name     string
	Conn     *websocket.Conn
	TimeTaken time.Duration
}

// Server represents the WebSocket server
type Server struct {
	Users    map[*websocket.Conn]*User
	Mutex    sync.Mutex
	Upgrader websocket.Upgrader
	Winner   *User
	StartTime time.Time
}

func NewServer() *Server {
	return &Server{
		Users: make(map[*websocket.Conn]*User),
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := s.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}
	defer conn.Close()

	var name string
	if err := conn.ReadJSON(&name); err != nil {
		log.Printf("Error reading JSON: %v", err)
		return
	}

	user := &User{Name: name, Conn: conn}
	s.Mutex.Lock()
	s.Users[conn] = user
	s.Mutex.Unlock()

	s.broadcastUsers()

	for {
		var msg string
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading JSON: %v", err)
			break
		}

		if msg == "time_in" {
			s.Mutex.Lock()
			user.TimeTaken = time.Since(s.StartTime)
			if s.Winner == nil || user.TimeTaken < s.Winner.TimeTaken {
				s.Winner = user
			}
			s.Mutex.Unlock()
		} else if msg == "reset" {
			s.Mutex.Lock()
			s.Winner = nil
			s.StartTime = time.Now()
			for _, u := range s.Users {
				u.TimeTaken = 0
			}
			s.Mutex.Unlock()
		}

		s.broadcastWinner()
	}

	s.Mutex.Lock()
	delete(s.Users, conn)
	s.Mutex.Unlock()
	s.broadcastUsers()
}

func (s *Server) broadcastUsers() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	userList := make([]string, 0, len(s.Users))
	for _, user := range s.Users {
		userList = append(userList, user.Name)
	}

	for _, user := range s.Users {
		if err := user.Conn.WriteJSON(userList); err != nil {
			log.Printf("Error writing JSON: %v", err)
		}
	}
}

func (s *Server) broadcastWinner() {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	winnerInfo := struct {
		Winner    string        `json:"winner"`
		TimeTaken time.Duration `json:"time_taken"`
	}{}

	if s.Winner != nil {
		winnerInfo.Winner = s.Winner.Name
		winnerInfo.TimeTaken = s.Winner.TimeTaken
	}

	for _, user := range s.Users {
		if err := user.Conn.WriteJSON(winnerInfo); err != nil {
			log.Printf("Error writing JSON: %v", err)
		}
	}
}

func main() {
	server := NewServer()
	http.HandleFunc("/ws", server.handleConnections)
	server.StartTime = time.Now()
	fmt.Println("WebSocket server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
