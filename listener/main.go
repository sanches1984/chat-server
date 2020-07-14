package main

import (
	"github.com/go-yaml/yaml"
	"github.com/sanches1984/chat-server/server"
	"io/ioutil"
	"log"
	"net/http"
)

type Config struct {
	Host    string `yaml:"host"`
	Port    string `yaml:"port"`
	Channel string `yaml:"channel"`
}

func main() {
	config, err := loadConfig()
	if err != nil {
		panic(err)
	}

	log.Println("Starting server:", config.Host, "/", config.Channel)
	srv, err := server.NewServer(config.Host)
	if err != nil {
		panic(err)
	}

	go srv.Run()
	http.HandleFunc("/chat/", func(w http.ResponseWriter, r *http.Request) {
		srv.Serve(w, r)
	})

	log.Println("Server started: listening", config.Port, "...")
	err = http.ListenAndServe(":"+config.Port, nil)
	if err != nil {
		panic(err)
	}
}

func loadConfig() (*Config, error) {
	bytes, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
