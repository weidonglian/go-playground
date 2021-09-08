package cache

import "errors"

var (
	ErrorEntryNotFound = errors.New("cache entry not found")
)

type Key string
type Value interface{}

type Cache interface {
	Get(key Key) (Value, error)
	Set(key Key, val Value) error
}
