package main

import (
	"chat/trace"
	"github.com/labstack/gommon/log"
	"net/http"
	"github.com/gorilla/websocket"
)

type room struct {
	forward chan []byte
	join chan *client
	leave chan *client
	clients map[*client]bool
	tracer trace.Tracer
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join: make(chan *client),
		leave: make(chan *client),
		clients: make(map[*client]bool),

	}
}

func (r *room) run() {
	for {
		select {
		case client := <- r.join:
			r.clients[client] = true
			r.tracer.Trace("新しいクライアントが参加しました")
		case client := <- r.leave:
			delete(r.clients, client)
			close(client.send)
			r.tracer.Trace("クライアントが退出しました")
		case msg := <- r.forward:
			r.tracer.Trace("メッセージを受信しました:", string(msg))
			for client := range r.clients {
				select {
				case client.send <- msg:
					// メッセージ送信
					r.tracer.Trace(" -- クライアントに送信されました")
				default:
					delete(r.clients, client)
					close(client.send)
					r.tracer.Trace("  -- 送信に失敗しました。クライアントをクリーンアップします。")
				}
			}
		}
	}
}

const (
	socketBufferSize = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP: ", err)
		return
	}
	client := &client{
		socket: socket,
		send: make(chan []byte, messageBufferSize),
		room: r,
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}