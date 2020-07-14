package server

import (
	"github.com/sanches1984/chat-server/model"
	"log"
)

type hub struct {
	storage    *storage
	clients    map[string]*Client
	broadcast  chan *model.Message
	register   chan *Client
	unregister chan *Client
	reconnect  *chan bool
}

func newHub(storage *storage, reconnect *chan bool) *hub {
	return &hub{
		storage:    storage,
		broadcast:  make(chan *model.Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		reconnect:  reconnect,
	}
}

func (h *hub) run() {
	for {
		select {
		case <-*h.reconnect:
			go h.closeAllClients()
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.processMessage(message)
		}
	}
}

func (h *hub) registerClient(client *Client) {
	log.Println("Register client:", client.username)
	err := h.storage.addSession(client.username)
	if err != nil {
		log.Println(err.Error())
		return
	}

	h.clients[client.username] = client
	h.processMessage(&model.Message{
		UserName: client.username,
		Type:     model.MessageEnter,
		Message:  "joined.",
	})
}

func (h *hub) unregisterClient(client *Client) {
	log.Println("Unregister client:", client.username)
	err := h.storage.deleteSession(client.username)
	if err != nil {
		log.Println(err.Error())
		return
	}

	h.processMessage(&model.Message{
		UserName: client.username,
		Type:     model.MessageExit,
		Message:  "left.",
	})

	if _, ok := h.clients[client.username]; ok {
		delete(h.clients, client.username)
		close(client.send)
	}
}

func (h *hub) processMessage(message *model.Message) {
	log.Println("Process message:", string(message.ToJSON()))

	switch message.Type {
	case model.MessagePublic:
		h.storage.updateSession(message.UserName)
		h.sendAll(message)
	case model.MessagePrivate:
		h.storage.updateSession(message.UserName)
		h.sendPersonal(&model.Message{
			UserName: message.To,
			Type:     model.MessagePrivate,
			Message:  message.Message,
		})
		h.sendPersonal(&model.Message{
			UserName: message.UserName,
			Type:     model.MessagePrivate,
			Message:  message.Message,
		})
	case model.MessageStat:
		list, err := h.getSessions()
		if err != nil {
			log.Println("Error:", err)
			return
		}

		h.sendPersonal(&model.Message{
			UserName: message.UserName,
			Type:     model.MessageStat,
			Message:  list.GetStatText(),
		})
	case model.MessageList:
		list, err := h.getSessions()
		if err != nil {
			log.Println("Error:", err)
			return
		}

		h.sendPersonal(&model.Message{
			UserName: message.UserName,
			Type:     model.MessageList,
			Message:  list.GetUsers(),
		})
	case model.MessageEnter:
		h.sendAll(message)
	case model.MessageExit:
		h.sendAll(message)
	}
}

func (h *hub) sendPersonal(message *model.Message) {
	client := h.clients[message.UserName]
	select {
	case client.send <- message.ToJSON():
	default:
		close(client.send)
		delete(h.clients, client.username)
	}
}

func (h *hub) sendAll(message *model.Message) {
	for _, client := range h.clients {
		select {
		case client.send <- message.ToJSON():
		default:
			close(client.send)
			delete(h.clients, client.username)
		}
	}
}

func (h *hub) getSessions() (model.SessionList, error) {
	userNames := make([]string, 0, len(h.clients))
	for client, _ := range h.clients {
		userNames = append(userNames, client)
	}

	return h.storage.getSessions(userNames)
}

func (h *hub) closeAllClients() {
	log.Println("Hub reconnect start...")
	for _, client := range h.clients {
		h.unregister <- client
	}
	log.Println("Hub reconnect complete.")
}
