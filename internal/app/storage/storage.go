package storage

import (
	"fmt"
	"sync"
)

var storage = sync.Map{}

func SetValue(key, value string) {
	storage.Store(key, value)
}

func GetValue(key string) (string, bool) {
	value, ok := storage.Load(key)
	return fmt.Sprint(value), ok
}
