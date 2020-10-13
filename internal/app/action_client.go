package app

import (
	"strconv"
)

func (c *Client) ActionUpdateName(message *Message) {
	name := GetString("display_name", message.Data)
	if name == "" || len(name) < 4 {
		c.Send(NewError("نام باید حداقل ۴ کلمه باشد"))
		return
	}
	c.SetDisplayName(name)
	// broadcast to everyone in the room that we have changed our name
	c.BroadcastRoom(NewMessage("update_name", JSON{
		"client_id": c.UID,
		"new_name":  name,
	}))
	c.SendMe()
}

// When user uses create_room action
// server sends them a unique id for
// a new room

func (c *Client) ActionCreateRoom(message *Message) {

	if !c.DisplayNameIsValid() {
		c.Send(NewError("نام نمایشی خود را اول وارد کنید"))
		return
	}

	room := NewRoom()
	c.Server.AddRoom(room)
	c.Send(NewMessage("room_created", JSON{
		"text":    "اتاق با موفقیت ساخته شد",
		"room_id": room.ID,
	}))
}

// when room is created users can now join
// the new room

func (c *Client) ActionJoinRoom(message *Message) {

	if !c.DisplayNameIsValid() {
		c.Send(NewError("نام نمایشی خود را اول وارد کنید"))
		return
	}

	uid := GetString("room_id", message.Data)
	if uid == "" {
		c.Send(NewError("اتاق پیدا نشد"))
		return
	}
	room := c.Server.GetRoom(uid)
	if room == nil {
		c.Send(NewError("اتاق پیدا نشد"))
		return
	}

	// if already in room
	if room.IsUserInRoom(c.ID()) {
		c.Send(NewMessage("join_room", JSON{
			"text": "وارد اتاق جدید میشوید",
			"room": room.Map(),
		}))
		return
	}

	if room.CountUsers() < 2 {
		room.AddClient(c)
		c.SetRoom(room)

		// check if game should be started
		if room.CountUsers() == 2 {
			room.StartGame()
		}

		c.Send(NewMessage("join_room", JSON{
			"text": "وارد اتاق جدید میشوید",
			"room": room.Map(),
		}))

	} else {
		c.Send(NewMessage("room_full", JSON{
			"text": "اتاق پر میباشد!",
		}))
	}

}

// Check if user is currently in a game or waiting in lobby
// This method only returns game status of a user
// Full status of the game is given by another method

func (c *Client) ActionGetGameStatus(message *Message) {
	status := "lobby"
	if c.GetGameStatus() == STATUS_INROOM {
		status = "inroom"
	}
	c.Send(NewMessage("room_status", JSON{
		"status": status,
	}))
}

// Broadcast a text message to room users
func (c *Client) ActionBroadcast(message *Message) {
	msg := GetString("text", message.Data)
	if msg == "" {
		return
	}
	if c.CurrentRoom == nil {
		return
	}

	c.CurrentRoom.Broadcast(NewMessage("chat", JSON{
		"text": msg,
		"from": c.GetDisplayName(),
	}), nil)
}

// Broadcast a text message to room users
func (c *Client) ActionSelect(message *Message) {
	i := GetString("i", message.Data)
	j := GetString("j", message.Data)

	if i == "" || j == "" {
		return
	}

	i_int, _ := strconv.Atoi(i)
	j_int, _ := strconv.Atoi(j)

	if i_int < 0 || i_int > 2 || j_int < 0 || j_int > 2 {
		return
	}

	if c.CurrentRoom == nil {
		return
	}

	if c.CurrentRoom.TurnUID() != c.UID {
		c.Send(NewMessage("not_your_turn", JSON{}))
		c.CurrentRoom.BroadcastUpdate()
		return
	}

	if !c.CurrentRoom.CanChangeIJ(i_int, j_int) {
		c.Send(NewMessage("already_set", JSON{}))
		c.CurrentRoom.BroadcastUpdate()
		return
	}
	c.CurrentRoom.Select(i_int, j_int)
	c.CurrentRoom.SwitchTurn()
	c.CurrentRoom.CheckGameHasEnded()
	c.CurrentRoom.BroadcastUpdate()
}

// Re game
func (c *Client) ActionRegame(message *Message) {
	if c.CurrentRoom == nil {
		return
	}

	c.CurrentRoom.AddToReGame(c.UID)

	if c.CurrentRoom.CanRegame() {

		newRoom := NewRoom()
		c.Server.AddRoom(newRoom)

		for _, cli := range c.CurrentRoom.GetClients() {
			newRoom.AddClient(cli)
			cli.SetRoom(newRoom)
			cli.Send(NewMessage("join_room", JSON{
				"text": "وارد اتاق جدید میشوید",
				"room": newRoom.Map(),
			}))
		}

		newRoom.StartGame()
		newRoom.BroadcastUpdate()

	} else {
		c.CurrentRoom.Broadcast(NewMessage("chat", JSON{
			"text": "درخواست بازی جدید",
			"from": c.GetDisplayName(),
		}), nil)
	}
}
