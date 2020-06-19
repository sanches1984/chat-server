package server

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sanches1984/chat-server/model"
	"log"
	"net/http"
	"strings"
	"time"
)

type ClientWS websocket.Conn

const writeWait = 500 * time.Microsecond

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type webSocketClient struct {
	conn     *ClientWS
	UserName string
}

func (c server) subscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	params := strings.Split(r.URL.Path, "/")
	userName := params[len(params)-1]
	if userName == "" {
		w.WriteHeader(400)
		return
	}
	if _, ok := c.hub.clients[userName]; ok {
		w.WriteHeader(403)
		log.Println("Username exists:", userName)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}

	clientWS := ClientWS(*conn)
	client := &webSocketClient{conn: &clientWS, UserName: userName}
	c.hub.register <- client
	c.hub.outgoing <- &model.Message{
		UserName: userName,
		Type:     model.MessageEnter,
		Message:  "joined...",
	}
}

func (cws *ClientWS) writeMessage(message []byte) error {
	ws := websocket.Conn(*cws)
	ws.SetWriteDeadline(time.Now().Add(writeWait))
	w, err := ws.NextWriter(websocket.TextMessage)
	if err != nil {
		return fmt.Errorf("Cant open websocket writer: %v", err)
	}
	defer w.Close()
	w.Write(message)
	return nil
}

func (cws *ClientWS) close() error {
	conn := websocket.Conn(*cws)
	return conn.Close()
}
