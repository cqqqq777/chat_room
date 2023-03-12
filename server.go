package services

import (
	"log"
	"net/http"
	"sync"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/websocket"
)

type ClientPool struct {
	users map[*Client]bool //在线则设置为true，离线删除
	Mutex sync.RWMutex
}

type Client struct {
	conn *websocket.Conn
	name string
}

var upgrader = websocket.HertzUpgrader{CheckOrigin: func(_ *app.RequestContext) bool {
	return true
}}

var (
	joinChan  = make(chan []byte)       //登入管道
	msgChan   = make(chan []byte, 1024) //消息管道
	leaveChan = make(chan []byte)       //登出管道
)

var clientPool = ClientPool{
	users: make(map[*Client]bool),
	Mutex: sync.RWMutex{},
}

func ServeWs(c *app.RequestContext) {
	//升级websocket
	err := upgrader.Upgrade(c, func(conn *websocket.Conn) {
		name, ok := c.Get("name")
		if !ok {
			log.Println("no username")
			return
		}
		//创建client实例
		var client = &Client{
			conn: conn,
			name: name.(string),
		}
		clientPool.join(client)
		//启动协程来读取用户发送的消息
		go client.readMsg()
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, "connect failed")
		log.Println(err)
		return
	}
}

func (c *ClientPool) join(client *Client) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	c.users[client] = true
	joinChan <- []byte("user:" + client.name + "join")
}

func (c *ClientPool) leave(client *Client) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	delete(c.users, client)
	leaveChan <- []byte("user:" + client.name + "leave")
}

func ListenAndServe() {
	for {
		//监控用户的登入登出与发消息
		//获取到消息后遍历用户组，将消息分别发给他们
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

func (c *Client) readMsg() {
	//读用户发来的消息
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			log.Println(err)
			continue
		}
		msgChan <- data
	}
}
