package redis

import "time"

type Service interface {
	// Set stores a key-value pair with an expiration time
	Set(key string, value any, expiration time.Duration) error

	// Get retrieves the value associated with the given key
	Get(key string) (string, error)

	// Delete removes the specified key and its value
	Delete(key string) error

	// Exists checks if a key exists in Redis
	Exists(key string) (bool, error)
}
