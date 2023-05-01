package storage

var storage = make(map[string]string)

func SetValue(key, value string) {
	storage[key] = value
}

func GetValue(key string) string {
	return storage[key]
}
