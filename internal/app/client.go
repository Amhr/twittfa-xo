package app

import (
	"OnlineXO/pkg/lock"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

type Client struct {
	send         chan Message
	conn         Connection
	DisplayName  string
	Server       *Server
	CurrentRoom  *Room
	UID          string
	Token        string
	Context      context.Context
	contextClose context.CancelFunc
	Status       string
	GameStatus   int
	closeOne     *sync.Once
}

const STATUS_LOBBY = 0
const STATUS_INROOM = 1

func NewClient(con Connection, s *Server) *Client {
	ctx, cn := context.WithCancel(context.Background())
	c := &Client{
		send:         make(chan Message, 100),
		conn:         con,
		Server:       s,
		CurrentRoom:  nil,
		UID:          uuid.New().String(),
		Token:        uuid.New().String(),
		Context:      ctx,
		contextClose: cn,
		Status:       "online",
		GameStatus:   STATUS_LOBBY,
		closeOne:     new(sync.Once),
	}
	c.Worker()
	c.SendMe()
	return c
}

func (c *Client) Map() JSON {
	return JSON{
		"uid":          c.UID,
		"display_name": c.DisplayName,
	}
}

func (c *Client) Send(msg *Message) {
	c.send <- *msg
}

func (c *Client) SetRoom(room *Room) {
	lock.GlobalLock.Get("client", c.ID()).Lock()
	c.CurrentRoom = room
	lock.GlobalLock.Get("client", c.ID()).Unlock()
}

func (c *Client) GetRoom() *Room {
	lock.GlobalLock.Get("client", c.ID()).Lock()
	defer lock.GlobalLock.Get("client", c.ID()).Unlock()
	return c.CurrentRoom
}

func (c *Client) BroadcastRoom(msg *Message) {
	if c.CurrentRoom != nil {
		c.CurrentRoom.Broadcast(msg, c)
	}
}

func (c *Client) SetDisplayName(name string) {
	lock.GlobalLock.Get("client", c.ID()).Lock()
	c.DisplayName = name
	lock.GlobalLock.Get("client", c.ID()).Unlock()
}

func (c *Client) GetDisplayName() string {
	lock.GlobalLock.Get("client", c.ID()).Lock()
	defer lock.GlobalLock.Get("client", c.ID()).Unlock()
	return c.DisplayName
}

func (c *Client) SetConnection(conn Connection) {
	lock.GlobalLock.Get("client", c.ID()).Lock()
	c.conn = conn
	lock.GlobalLock.Get("client", c.ID()).Unlock()
}

func (c *Client) DisplayNameIsValid() bool {
	lock.GlobalLock.Get("client", c.ID()).Lock()
	defer lock.GlobalLock.Get("client", c.ID()).Unlock()
	return c.DisplayName != "" && len(c.DisplayName) > 4
}

func (c *Client) SetStatus(status string) {
	lock.GlobalLock.Get("client", c.ID(), "status").Lock()
	c.Status = status
	lock.GlobalLock.Get("client", c.ID(), "status").Unlock()
}

func (c *Client) GetStatus() string {
	lock.GlobalLock.Get("client", c.ID(), "status").Lock()
	defer lock.GlobalLock.Get("client", c.ID(), "status").Unlock()
	return c.Status
}

func (c *Client) IsOnline() bool {
	return c.GetStatus() != "offline"
}

func (c *Client) SendMe() {
	room_id := ""
	room := c.GetRoom()
	if room != nil {
		room_id = room.ID
	}

	c.Send(NewMessage("me", JSON{
		"token":       c.Token,
		"client":      c.Map(),
		"game_status": c.GetGameStatus(),
		"public_id":   c.UID,
		"room_id":     room_id,
	}))
}

func (c *Client) GetGameStatus() int {
	lock.GlobalLock.Get("client", c.ID(), "status").Lock()
	defer lock.GlobalLock.Get("client", c.ID(), "status").Unlock()
	return c.GameStatus
}

func (c *Client) SetGameStatus(status int) {
	lock.GlobalLock.Get("client", c.ID(), "status").Lock()
	defer lock.GlobalLock.Get("client", c.ID(), "status").Unlock()
	c.GameStatus = status
}

// WORKERS
// we have a worker to listen on new messages
// and another worker to send new messages

// This method listens to receive new messages
func (c *Client) listen() {

	alive := true
	for alive {
		select {
		case <-c.Context.Done():
			c.Close()
			alive = false
		default:
			_, msg, err := c.conn.ReadMessage()
			if err != nil {
				c.contextClose()
				c.Close()
				alive = false
				return
			}
			var jmsg JSON

			err = json.Unmarshal(msg, &jmsg)

			if err != nil {
				fmt.Println("Message ")
				fmt.Println(msg)
				//todo: Handle unusual messages
				log.Printf("Unusual Message From : %s , %s", c.conn.RemoteAddr(), string(msg))
				continue
			}

			// message is json , lets convert it into Message
			message, err := MessageFromJson(jmsg)

			if err != nil {
				log.Printf("Unusual Json %v : %s , %s", err, c.conn.RemoteAddr(), string(msg))
				continue
			}

			c.handle(message)
		}

	}
}

// Send Worker
func (c *Client) sender() {
	alive := true
	for alive {
		select {
		case <-c.Context.Done():
			alive = false
		default:
			msg := <-c.send
			js, err := json.Marshal(msg.Map())
			if err != nil {
				log.Printf("Problem while encoding system message, %s , %v", msg, err)
				continue
			}
			err = c.conn.WriteMessage(websocket.TextMessage, js)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

// Runs workers
// 1. Listen Worker which listens for new messages.
// 2. Send Worker which listens on a channel to receive new messages
// 		- this worker is the only way to send message
func (c *Client) Worker() {
	go c.sender()
	go c.listen()
}

// get string id of client
func (c *Client) ID() string {
	return c.Token
}

// closing client

func (c *Client) Close() {
	c.SetStatus("offline")
	c.BroadcastRoom(NewMessage("offline", JSON{
		"client": c.Map(),
	}))
	c.closeOne.Do(func() {
		close(c.send)
	})
	_ = c.conn.Close()

}

func (c *Client) Reconnect() {
	c.SetStatus("online")
	c.BroadcastRoom(NewMessage("online", JSON{
		"client": c.Map(),
	}))
	c.send = make(chan Message)
	ctx, can := context.WithCancel(context.Background())
	c.closeOne = new(sync.Once)
	c.Context = ctx
	c.contextClose = can
	c.Worker()
	c.SendMe()
}

//todo: Ensuring all goroutine will die after client get disconnected
