package lock

import (
	"strings"
	"sync"
)

type Lock struct {
	global sync.Mutex
	locks  map[string]*sync.Mutex
}

func (l *Lock) Get(keys ...string) *sync.Mutex {
	key := strings.Join(keys, "_")
	l.global.Lock()
	lock, exists := l.locks[key]
	if !exists {
		lock = &sync.Mutex{}
		l.locks[key] = lock
	}
	l.global.Unlock()
	return lock
}

var GlobalLock = &Lock{
	global: sync.Mutex{},
	locks:  map[string]*sync.Mutex{},
}
