package cache

import (
	mocks "github.com/weidonglian/go-playground/mocks/cache"
	"testing"
)

func play(cache Cache, t *testing.T) {
	t.Log(cache.Get("hello"))
	cache.Set("x", "y")
}

func TestCache(t *testing.T) {
	mockedCache := new(mocks.Cache)
	mockedCache.On("Get", "hello").Return("world", nil)
	mockedCache.On("Set", "x", "y").Return(nil)
	play(mockedCache, t)
	mockedCache.AssertExpectations(t)
}
