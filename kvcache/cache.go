package kvcache

import (
	"fmt"
	"time"
)

var (
	ErrNoDataToSet         = fmt.Errorf("empty map received")
	ErrNoKeysToRetrieve    = fmt.Errorf("empty keys slice received")
	ErrKeyNotExist         = fmt.Errorf("key does not exist")
	ErrStorageAlreadyEmpty = fmt.Errorf("clearing empty caching storage")
	ErrInvalidDuration     = fmt.Errorf("time duration must be greater than 0")
)

type DB[K comparable, T any] interface {
	Get(key K) (T, bool)
	Set(key K, val T) error
	GetMany(keys []K) (map[K]T, error)
	SetMany(pairs map[K]T) error
	Exists(key K) bool
	Delete(key K) error
	Clear() error
	Size() int
	SetTTL(ttl, checkupTimeout time.Duration) error
	Stop()
}

// entry represents single cache value.
type entry[T any] struct {
	payload T
	created time.Time
}

// db represents a simple key-value caching storage.
type db[K comparable, T any] struct {
	storage        map[K]entry[T]
	ttl            time.Duration
	checkupTimeout time.Duration
	done           chan struct{}
}

// Get retrieves a value from the caching DB storage by key.
// Returns the value associated with the key, and a boolean indicating
// whether the key was found in the storage.
func (d *db[K, T]) Get(key K) (T, bool) {
	en, ok := d.storage[key]

	return en.payload, ok
}

// Set adds or updates a key-value pair in the caching DB storage.
// Returns an error if the operation fails.
func (d *db[K, T]) Set(key K, val T) error {
	d.storage[key] = entry[T]{
		payload: val,
		created: time.Now(),
	}

	return nil
}

// GetMany accepts a slice of keys and returns a map of occurrences from the caching DB storage.
// Returns an error if the operation fails.
func (d *db[K, T]) GetMany(keys []K) (map[K]T, error) {
	if len(keys) == 0 {
		return nil, ErrNoKeysToRetrieve
	}

	res := make(map[K]T)
	for _, key := range keys {
		if val, ok := d.storage[key]; ok {
			res[key] = val.payload
		}
	}

	return res, nil
}

// SetMany atomically sets multiple key-value pairs in the caching DB storage.
// Returns an error if the operation fails.
func (d *db[K, T]) SetMany(pairs map[K]T) error {
	if len(pairs) == 0 {
		return ErrNoDataToSet
	}

	for k, v := range pairs {
		d.storage[k] = entry[T]{
			payload: v,
			created: time.Now(),
		}
	}

	return nil
}

// Exists returns whether the key is present in the caching DB storage.
func (d *db[K, T]) Exists(key K) bool {
	_, ok := d.storage[key]
	return ok
}

// Delete deletes an entry with the specified key from the caching DB storage.
// If the key does not exist, ErrKeyNotExist is returned.
func (d *db[K, T]) Delete(key K) error {
	if !d.Exists(key) {
		return ErrKeyNotExist
	}

	delete(d.storage, key)
	return nil
}

// Clear clears the caching DB storage.
// Returns ErrStorageAlreadyEmpty if calling on an empty caching storage.
func (d *db[K, T]) Clear() error {
	if len(d.storage) == 0 {
		return ErrStorageAlreadyEmpty
	}

	clear(d.storage)
	return nil
}

// Size returns the number of key-value pairs stored in the caching DB storage.
func (d *db[K, T]) Size() int {
	return len(d.storage)
}

// SetTTL sets the time-to-live (TTL) for caching entries.
// ttl is the duration after which the entry will be automatically removed.
// checkupTimeout is the duration between checks for expired entries.
// If ttl or checkupTimeout are <= 0, ErrInvalidDuration is returned.
// This function starts a goroutine that periodically checks for expired entries and removes them.
func (d *db[K, T]) SetTTL(ttl, checkupTimeout time.Duration) error {
	if ttl <= 0 || checkupTimeout <= 0 {
		return ErrInvalidDuration
	}

	ticker := time.NewTicker(checkupTimeout)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				for k, v := range d.storage {
					if time.Since(v.created) > d.ttl {
						delete(d.storage, k)
					}
				}
			case <-d.done:
				return
			}
		}
	}()

	return nil
}

func (d *db[K, T]) Stop() {
	close(d.done)
}

func New[K comparable, T any]() DB[K, T] {
	return &db[K, T]{
		storage: make(map[K]entry[T]),
		done:    make(chan struct{}),
	}
}
