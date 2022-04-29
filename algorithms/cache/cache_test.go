package cache

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/weidonglian/go-playground/mocks"
)

func play(cache Cache, t *testing.T) {
	t.Log(cache.Get("hello"))
	cache.Set("x", "y")
}

func TestCache(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockedCache := mocks.NewMockCache(mockCtrl)
	mockedCache.EXPECT().Get("hello").Return("world", nil)
	mockedCache.EXPECT().Set("x", "y").Return(nil)
	play(mockedCache, t)
}
