package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/websocket"
)

type Client struct {
	conn *websocket.Conn
	name string
}

func NewClient(c *websocket.Conn) *Client {
	return &Client{conn: c}
}

func (client *Client) readMsg() {
	for {
		_, p, err := client.conn.ReadMessage()
		if err != nil {
			log.Println("read message failed err:", err)
			continue
		}
		fmt.Println(string(p))
	}
}

func (client *Client) sendMsg() {
	for {
		reader := bufio.NewReader(os.Stdin)
		bytes, err := reader.ReadBytes('\n')
		if err != nil {
			log.Println("send message failed err:", err)
			continue
		}
		err = client.conn.WriteMessage(1, bytes[:len(bytes)-1])
		if err != nil {
			log.Println("send message failed err:", err)
			continue
		}
	}
}

func main() {
	url := "ws://localhost:8080/ws"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := NewClient(conn)
	go client.readMsg()
	go client.sendMsg()
	select {}
}
