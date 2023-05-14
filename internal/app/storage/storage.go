package storage

import (
	"fmt"
	"sync"
)

type storage struct {
	mu    sync.Mutex
	links map[string]string
}

func (c *storage) SetValue(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.links[key] = value
}

var s = storage{
	links: map[string]string{},
}

func SetValue(key, value string) {
	s.SetValue(key, value)
}

func GetValue(key string) (string, bool) {
	value, ok := s.links[key]
	return fmt.Sprint(value), ok
}
