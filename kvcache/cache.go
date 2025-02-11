package kvcache

import "fmt"

var (
	ErrNoDataToSet         = fmt.Errorf("empty map received")
	ErrNoKeysToRetrieve    = fmt.Errorf("empty keys slice received")
	ErrKeyNotExist         = fmt.Errorf("key does not exist")
	ErrStorageAlreadyEmpty = fmt.Errorf("clearing empty caching storage")
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
}

type db[K comparable, T any] struct {
	storage map[K]T
}

// Get retrieves a value from the caching DB storage by key.
// Returns the value associated with the key, and a boolean indicating
// whether the key was found in the storage.
func (d *db[K, T]) Get(key K) (T, bool) {
	val, ok := d.storage[key]

	return val, ok
}

// Set adds or updates a key-value pair in the caching DB storage.
// Returns an error if the operation fails.
func (d *db[K, T]) Set(key K, val T) error {
	d.storage[key] = val

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
			res[key] = val
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
		d.storage[k] = v
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

func New[K comparable, T any]() DB[K, T] {
	return &db[K, T]{
		storage: make(map[K]T),
	}
}
