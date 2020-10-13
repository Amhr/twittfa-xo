package app

import (
	"sync"
)

type Server struct {
	Clients    map[string]*Client
	Rooms      map[string]*Room
	globalLock sync.Mutex
}

func NewServer() *Server {
	return &Server{Clients: map[string]*Client{}, Rooms: map[string]*Room{}}
}

func (s *Server) AddClient(client *Client) {
	s.globalLock.Lock()
	s.Clients[client.ID()] = client
	s.globalLock.Unlock()
}

func (s *Server) GetClientByToken(token string) *Client {
	s.globalLock.Lock()
	c, e := s.Clients[token]
	s.globalLock.Unlock()
	if !e {
		return nil
	}
	return c
}

func (s *Server) AddRoom(room *Room) {
	s.globalLock.Lock()
	s.Rooms[room.ID] = room
	s.globalLock.Unlock()
}

func (s *Server) GetRoom(uid string) *Room {
	s.globalLock.Lock()
	room, exists := s.Rooms[uid]
	s.globalLock.Unlock()
	if !exists {
		return nil
	}
	return room
}
