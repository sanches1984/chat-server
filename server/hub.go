package server

import (
	"context"
	"github.com/sanches1984/chat-server/model"
	"log"
)

type hub struct {
	storage    *storage
	register   chan *webSocketClient
	unregister chan *webSocketClient
	incoming   chan *model.Message
	outgoing   chan *model.Message
	reconnect  *chan bool
	clients    map[string]*webSocketClient
}

func newHub(storage *storage, reconnect *chan bool) *hub {
	return &hub{
		storage:    storage,
		register:   make(chan *webSocketClient),
		unregister: make(chan *webSocketClient),
		incoming:   make(chan *model.Message),
		outgoing:   make(chan *model.Message),
		reconnect:  reconnect,
		clients:    make(map[string]*webSocketClient),
	}
}

func (h *hub) run(ctx context.Context) {
	for {
		select {
		case <-*h.reconnect:
			log.Println("Hub reconnect start...")
			go h.closeAllClients(ctx)
		case client := <-h.register:
			log.Println("Register client:", client.UserName)
			h.clients[client.UserName] = client
			err := h.storage.addSession(client.UserName)
			if err != nil {
				log.Println(ctx, "Can't set user:", err)
			}
		case client := <-h.unregister:
			log.Println("Unregister client:", client.UserName)
			err := client.conn.close()
			if err != nil {
				log.Println("Can't close websocket:", err)
			}
			err = h.storage.deleteSession(client.UserName)
			if err != nil {
				log.Println("Can't drop session:", err)
			}
			if _, ok := h.clients[client.UserName]; ok {
				delete(h.clients, client.UserName)
			}
		case message := <-h.outgoing:
			log.Println("Send message:", message)
			go h.sendMessage(ctx, message)
		case message := <-h.incoming:
			log.Println("Recieve message:", message)
			go h.sendMessage(ctx, message)
		}
	}
}

func (h *hub) closeAllClients(ctx context.Context) {
	for _, client := range h.clients {
		h.unregister <- client
	}
	log.Println("Hub reconnect complete")
}

func (h *hub) sendMessage(ctx context.Context, message *model.Message) {
	for _, c := range h.clients {
		err := c.conn.writeMessage(message.ToJSON())
		if err != nil {
			log.Println("Write message error:", err)
		}
	}
}
