package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/sanches1984/chat-server/model"
	"log"
	"time"
)

const pubSubReconnectTimeoutSecond = 30

func (c server) pubSubListener(ctx context.Context) {
	for {
		pubSubConn, err := pubSubConnectAndSubscribe(c.host, c.channel)
		if err != nil {
			log.Println("Subscribe error:", err.Error())
			time.Sleep(pubSubReconnectTimeoutSecond * time.Second)
			continue
		}
	ListenLoop:
		for {
			switch v := pubSubConn.Receive().(type) {
			case redis.Message:
				log.Println("Receive message:", string(v.Data))
				message, err := getMessage(v.Data)
				if err != nil {
					log.Println("Message parse error:", err)
				}
				c.hub.outgoing <- message
			case error:
				log.Println(v.Error(), "\nWait and reconnect...")
				pubSubConn.Close()
				time.Sleep(pubSubReconnectTimeoutSecond * time.Second)
				break ListenLoop
			}
		}
	}
}

func PubSubConnect(host string) (redis.PubSubConn, error) {
	redisConn, err := redis.Dial("tcp", host)
	if err != nil {
		return redis.PubSubConn{}, fmt.Errorf("Cannot connect to redis: " + err.Error())
	}
	return redis.PubSubConn{Conn: redisConn}, nil
}

func pubSubConnectAndSubscribe(host, channel string) (redis.PubSubConn, error) {
	pubSubConn, err := PubSubConnect(host)
	if err != nil {
		return pubSubConn, err
	}
	err = pubSubConn.Subscribe(channel)
	if err != nil {
		return redis.PubSubConn{}, fmt.Errorf("Cannot redis subscribe: " + err.Error())
	}
	return pubSubConn, nil
}

func getMessage(data []byte) (*model.Message, error) {
	var msg model.Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
