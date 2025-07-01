package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
	From    string      `json:"from,omitempty"`
}

type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	server *Server
}

type Server struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mutex.Lock()
			s.clients[client] = true
			s.mutex.Unlock()
			log.Printf("Client connected. Total clients: %d", len(s.clients))

		case client := <-s.unregister:
			s.mutex.Lock()
			if _, ok := s.clients[client]; ok {
				delete(s.clients, client)
				close(client.send)
			}
			s.mutex.Unlock()
			log.Printf("Client disconnected. Total clients: %d", len(s.clients))

		case message := <-s.broadcast:
			s.mutex.RLock()
			for client := range s.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(s.clients, client)
				}
			}
			s.mutex.RUnlock()
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.server.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		c.server.broadcast <- message
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

func HandleWebSocket(server *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		client := &Client{
			conn:   conn,
			send:   make(chan []byte, 256),
			server: server,
		}

		client.server.register <- client

		go client.WritePump()
		go client.ReadPump()
	}
}

func main() {
	wsServer := NewServer()
	go wsServer.Run()

	fs := http.FileServer(http.Dir("public"))
	http.Handle("/", fs)

	http.HandleFunc("/ws", HandleWebSocket(wsServer))

	port := "8069"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	fmt.Printf("Serving on http://localhost:%s\n", port)
	fmt.Printf("WebSocket server available at ws://localhost:%s/ws\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
