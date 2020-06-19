package server

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"net"
	"sync"
	"time"
)

const prefix = "skt"

type storage struct {
	sync.Mutex
	host      string
	conn      *redis.Conn
	reconnect *chan bool
}

// Подключиться к хранилищу
func newStorage(reconnect *chan bool, host string) (*storage, error) {
	storage := &storage{reconnect: reconnect, host: host}
	err := storage.doConnect()
	if err != nil {
		return nil, err
	}
	return storage, nil
}

// Подключение
func (s *storage) doConnect() error {
	redisConn, err := redis.Dial("tcp", s.host)
	if err != nil {
		return fmt.Errorf("Cannot connect to redis: %v", err)
	}
	s.conn = &redisConn
	return nil
}

// Переподключение
func (s *storage) doReconnect() error {
	if s.conn != nil {
		_ = (*s.conn).Close()
	}
	s.conn = nil
	err := s.doConnect()
	if err == nil {
		*s.reconnect <- true
	}
	return err
}

// Выполняем команду, если сетевые проблемы, то реконнект и выполняем еще раз
func (s *storage) do(commandName string, args ...interface{}) (interface{}, error) {
	if s.conn == nil {
		return nil, fmt.Errorf("Redis disconnected")
	}
	firstAttemp := true
	s.Lock()
TryAgain:
	reply, err := (*s.conn).Do(commandName, args...)
	if firstAttemp && err != nil {
		if _, ok := err.(*net.OpError); ok {
			err = s.doReconnect()
			if err == nil {
				firstAttemp = false
				goto TryAgain
			}
		}
	}
	s.Unlock()
	return reply, err
}

// Добавляет запись о сессии юзера
func (s *storage) addSession(userName string) error {
	_, err := s.do("SET", userName, time.Now().String())
	return err
}

// Удаляет сессию юзера
func (s *storage) deleteSession(userName string) error {
	_, err := s.do("DEL", userName)
	return err
}
