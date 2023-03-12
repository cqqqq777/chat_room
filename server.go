package services

import (
	"log"
	"net/http"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"
)

type ClientPool struct {
	users map[*client]bool //在线则设置为true，离线删除
	Mutex sync.RWMutex
}

type client struct {
	conn *websocket.Conn
	name string
}

var upgrader = websocket.HertzUpgrader{CheckOrigin: func(_ *app.RequestContext) bool {
	return true
}}

var (
	joinChan  = make(chan []byte)
	msgChan   = make(chan []byte, 1024)
	leaveChan = make(chan []byte)
)

var clientPool = ClientPool{
	users: make(map[*client]bool),
	Mutex: sync.RWMutex{},
}

func ServeWs(c *app.RequestContext) {
	err := upgrader.Upgrade(c, func(conn *websocket.Conn) {
		name, ok := c.Get("name")
		if !ok {
			log.Println("no username")
			return
		}
		var client = &client{
			conn: conn,
			name: name.(string),
		}
		clientPool.join(client)

	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, "connect failed")
		log.Println(err)
		return
	}
}

func (c *ClientPool) join(client *client) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.users[client] = true
	joinChan <- []byte("user:" + client.name + "join")
}

func (c *ClientPool) leave(client *client) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	delete(c.users, client)
	leaveChan <- []byte("user:" + client.name + "leave")
}

func ListenAndServe() {
	for {
		select {
		case join := <-joinChan:
			for k, _ := range clientPool.users {
				err := k.conn.WriteMessage(1, join)
				if err != nil {
					log.Println(err)
				}
			}
		case msg := <-msgChan:
			for k, _ := range clientPool.users {
				err := k.conn.WriteMessage(1, msg)
				if err != nil {
					log.Println(err)
				}
			}
		case leave := <-leaveChan:
			for k, _ := range clientPool.users {
				err := k.conn.WriteMessage(1, leave)
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func (c *client) readMsg() {
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			continue
		}
		msgChan <- data
	}
}
