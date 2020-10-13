package app

import (
	"OnlineXO/pkg/lock"
	"github.com/google/uuid"
	"math/rand"
)

type Room struct {
	Clients map[string]*Client
	ID      string
	Status  int
	Board   [][]string
	Turn    string
	x_id    string
	o_id    string
	winner  string
	reGame  []string
}

const ROOM_CREATED = 0
const ROOM_STARTED = 1
const ROOM_ENDED = 2

func NewRoom() *Room {
	return &Room{
		Clients: map[string]*Client{},
		ID:      uuid.New().String(),
		Status:  ROOM_CREATED,
		Board:   [][]string{},
		Turn:    "",
		x_id:    "",
		o_id:    "",
		winner:  "",
		reGame:  []string{},
	}
}

func (r *Room) Map() JSON {
	return JSON{
		"id":      r.ID,
		"clients": r.GetClientsMap(),
		"status":  r.GetStatus(),
		"board":   r.Board,
		"turn":    r.Turn,
		"x_id":    r.x_id,
		"o_id":    r.o_id,
		"winner":  r.winner,
	}
}

func (r *Room) GetStatus() int {
	lock.GlobalLock.Get("room", r.ID).Lock()
	defer lock.GlobalLock.Get("room", r.ID).Unlock()
	return r.Status
}

func (r *Room) SetStatus(st int) {
	lock.GlobalLock.Get("room", r.ID).Lock()
	defer lock.GlobalLock.Get("room", r.ID).Unlock()
	r.Status = st
}

func (r *Room) AddClient(c *Client) {
	lock.GlobalLock.Get("room", r.ID).Lock()
	defer lock.GlobalLock.Get("room", r.ID).Unlock()
	r.Clients[c.ID()] = c
	c.SetGameStatus(STATUS_INROOM)
}

func (r *Room) IsUserInRoom(uid string) bool {
	lock.GlobalLock.Get("room", r.ID).Lock()
	defer lock.GlobalLock.Get("room", r.ID).Unlock()
	_, e := r.Clients[uid]
	return e
}

func (r *Room) StartGame() {
	r.SetStatus(ROOM_STARTED)
	lock.GlobalLock.Get("room", r.ID).Lock()
	xRandIndex := rand.Intn(100)
	if xRandIndex%2 == 0 {
		xRandIndex = 1
	} else {
		xRandIndex = 0
	}
	i := 0
	for _, c := range r.Clients {
		if i == xRandIndex {
			r.x_id = c.UID
		} else {
			r.o_id = c.UID
		}
		i++
	}

	for i := 0; i < 3; i++ {
		var s []string
		for j := 0; j < 3; j++ {
			s = append(s, "")
		}
		r.Board = append(r.Board, s)
	}
	r.Turn = "x"
	lock.GlobalLock.Get("room", r.ID).Unlock()

	r.Broadcast(NewMessage("game_started", JSON{}), nil)
	r.BroadcastUpdate()
}

func (r *Room) AddToReGame(uid string) {
	lock.GlobalLock.Get("room", r.ID).Lock()
	defer lock.GlobalLock.Get("room", r.ID).Unlock()
	for _, id := range r.reGame {
		if id == uid {
			return
		}
	}
	r.reGame = append(r.reGame, uid)
}

func (r *Room) CanRegame() bool {
	lock.GlobalLock.Get("room", r.ID).Lock()
	defer lock.GlobalLock.Get("room", r.ID).Unlock()
	return len(r.reGame) == 2
}

func (r *Room) TurnUID() string {
	lock.GlobalLock.Get("room", r.ID).Lock()
	defer lock.GlobalLock.Get("room", r.ID).Unlock()
	if r.Turn == "x" {
		return r.x_id
	}
	return r.o_id
}

func (r *Room) CanChangeIJ(i, j int) bool {
	lock.GlobalLock.Get("room", r.ID).Lock()
	defer lock.GlobalLock.Get("room", r.ID).Unlock()

	if i < 0 || i > 2 || j < 0 || j > 3 {
		return false
	}
	current := r.Board[i][j]
	return current == ""
}

func (r *Room) SwitchTurn() {
	lock.GlobalLock.Get("room", r.ID).Lock()
	defer lock.GlobalLock.Get("room", r.ID).Unlock()

	if r.Turn == "x" {
		r.Turn = "o"
	} else {
		r.Turn = "x"
	}
}

func (r *Room) Select(i, j int) {
	lock.GlobalLock.Get("room", r.ID).Lock()
	defer lock.GlobalLock.Get("room", r.ID).Unlock()
	r.Board[i][j] = r.Turn
}

func (r *Room) BroadcastUpdate() {
	r.Broadcast(NewMessage("update_room", JSON{
		"room": r.Map(),
	}), nil)
}

func (r *Room) CountUsers() int {
	return len(r.GetClients())
}

// Broadcast a Message to all clients
func (r *Room) Broadcast(message *Message, skip *Client) {
	for _, c := range r.Clients {
		if skip != nil && c.ID() == skip.ID() || !c.IsOnline() {
			continue
		}
		c.Send(message)
	}
}

// Get list of clients in room

func (r *Room) GetClients() map[string]*Client {
	lock.GlobalLock.Get("room", r.ID).Lock()
	c := r.Clients
	lock.GlobalLock.Get("room", r.ID).Unlock()
	return c
}

func (r *Room) CheckGameHasEnded() bool {
	lock.GlobalLock.Get("room", r.ID).Lock()
	defer lock.GlobalLock.Get("room", r.ID).Unlock()
	for i := 0; i < 3; i++ {
		if r.Board[i][0] == r.Board[i][1] && r.Board[i][0] == r.Board[i][2] && r.Board[i][0] != "" {
			r.Status = ROOM_ENDED
			r.winner = r.Board[i][0]
			return true
		}
	}

	for i := 0; i < 3; i++ {
		if r.Board[0][i] == r.Board[1][i] && r.Board[0][i] == r.Board[2][i] && r.Board[0][i] != "" {
			r.winner = r.Board[0][i]
			r.Status = ROOM_ENDED
			return true
		}
	}

	if r.Board[0][0] == r.Board[1][1] && r.Board[0][0] == r.Board[2][2] && r.Board[0][0] != "" {
		r.winner = r.Board[0][0]
		r.Status = ROOM_ENDED
		return true
	}

	if r.Board[2][0] == r.Board[1][1] && r.Board[2][0] == r.Board[0][2] && r.Board[2][0] != "" {
		r.winner = r.Board[2][0]
		r.Status = ROOM_ENDED
		return true
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if r.Board[i][j] == "" {
				return false
			}
		}
	}

	r.Status = ROOM_ENDED
	return true

}

func (r *Room) GetClientsMap() []JSON {
	c := r.GetClients()
	usrs := make([]JSON, 0)
	for _, u := range c {
		usrs = append(usrs, u.Map())
	}
	return usrs
}
