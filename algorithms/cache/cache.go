package cache

import "errors"

var (
	ErrorEntryNotFound = errors.New("cache entry not found")
)

type Cache interface {
	Get(key string) (interface{}, error)
	Set(key string, val interface{}) error
}
