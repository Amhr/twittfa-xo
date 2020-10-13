package app

func (c *Client) handle(message *Message) {
	switch message.Action {
	case "update_name":
		c.ActionUpdateName(message)
	case "create_room":
		c.ActionCreateRoom(message)
	case "join_room":
		c.ActionJoinRoom(message)
	case "broadcast":
		c.ActionBroadcast(message)
	case "select":
		c.ActionSelect(message)
	case "regame":
		c.ActionRegame(message)
	}

}
