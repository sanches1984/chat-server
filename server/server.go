package server

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
)

type server struct {
	storage *storage
	hub     *hub
	conn    *websocket.Conn
	host    string
	channel string
	send    chan []byte
}

type IServer interface {
	Serve(w http.ResponseWriter, r *http.Request)
	Run()
}

func NewServer(host string) (IServer, error) {
	reconnectChan := make(chan bool)
	storage, err := newStorage(&reconnectChan, host)
	if err != nil {
		return nil, err
	}

	hub := newHub(storage, &reconnectChan)
	return &server{
		storage: storage,
		hub:     hub,
	}, nil
}

func (c *server) Serve(w http.ResponseWriter, r *http.Request) {
	client := c.subscribe(w, r)
	if client == nil {
		log.Println("Client empty")
		return
	}
	go client.writePump()
	go client.readPump()
}

func (c *server) Run() {
	c.hub.run()
}

func (c *server) subscribe(w http.ResponseWriter, r *http.Request) *Client {
	params := strings.Split(r.URL.Path, "/")
	userName := params[len(params)-1]
	if userName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return nil
	}
	if _, ok := c.hub.clients[userName]; ok {
		w.WriteHeader(http.StatusForbidden)
		log.Println("Username exists:", userName)
		return nil
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	client := &Client{username: userName, hub: c.hub, conn: conn, send: make(chan []byte, 1024)}
	client.hub.register <- client
	return client
}
