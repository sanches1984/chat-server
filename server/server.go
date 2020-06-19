package server

import (
	"context"
	"net/http"
)

type server struct {
	storage *storage
	hub     *hub
	host    string
	channel string
}

type IServer interface {
	RunProcessor(ctx context.Context)
	RunSubListener(ctx context.Context)
	Serve(ctx context.Context, w http.ResponseWriter, r *http.Request)
}

func NewServer(host, channel string) (IServer, error) {
	reconnectChan := make(chan bool)
	storage, err := newStorage(&reconnectChan, host)
	if err != nil {
		return nil, err
	}

	hub := newHub(storage, &reconnectChan)
	return &server{
		storage: storage,
		hub:     hub,
		host:    host,
		channel: channel,
	}, nil
}

func (c server) RunProcessor(ctx context.Context) {
	c.hub.run(ctx)
}

func (c server) RunSubListener(ctx context.Context) {
	c.pubSubListener(ctx)
}

func (c server) Serve(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	c.subscribe(ctx, w, r)
	go c.RunProcessor(ctx)
	go c.RunSubListener(ctx)
}

//func (c server) Publish(message *model.Message) error {
//	log.Println("Publish message:", string(message.ToJSON()))
//	_, err := c.hub.Conn.Do("PUBLISH", c.channel, message.ToJSON())
//	return err
//}
