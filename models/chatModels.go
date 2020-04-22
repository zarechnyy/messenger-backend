package models

import (
	"github.com/gorilla/websocket"
)

type Room struct {
	ClientA Client
	ClientB Client
	ChatChannel chan *websocket.Conn `json:"-"`
}

type Client struct {
	User *User
	Ws *websocket.Conn               `json:"-"`
}
