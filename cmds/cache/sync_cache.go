package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gammazero/deque"
)

type Tile string

type TileRequest struct {
	zoom int
	x, y int
}

type TileResponse struct {
	tile     Tile
	reqTaken time.Duration
}

// emulate how to fetch the tile from a remote data server.
func FetchTile(zoom, x, y int) Tile {
	time.Sleep(5 * time.Second)
	return Tile(fmt.Sprintf("%d_%d_%d", zoom, x, y))
}

func HandleTileRequest(ctx context.Context, req TileRequest) (*TileResponse, error) {
	if ctx.Err() == context.Canceled {
		return nil, context.Canceled
	}
	start := time.Now()
	tile := getCachedOrFetchTile(req.zoom, req.x, req.y)
	return &TileResponse{
		tile:     tile,
		reqTaken: time.Since(start),
	}, nil
}

type TileCacheKey TileRequest

type TileCache interface {
	Get(key TileCacheKey) (Tile, bool)
	Set(key TileCacheKey, tile Tile)
	Close() error
}

// sync cache
type syncTileCache struct {
	capacity int
	hashmap  map[TileCacheKey]Tile // tileRequest with three ints are cheaper than string that requires heap allocation
	deque    deque.Deque           // used to keep track the eviction order and using a deque.Deque for a ring-buffer implmentation
	mutex    sync.RWMutex
	closec   chan struct{}
}

var _ TileCache = (*syncTileCache)(nil)

func (sc *syncTileCache) Get(key TileCacheKey) (Tile, bool) {
	sc.mutex.RLock()
	if v, ok := sc.hashmap[key]; ok {
		sc.mutex.RUnlock()
		return v, true
	}
	sc.mutex.RUnlock()
	return "", false
}

func (sc *syncTileCache) Set(key TileCacheKey, tile Tile) {
	sc.mutex.Lock()
	sc.hashmap[key] = tile
	sc.deque.PushBack(key)
	sc.mutex.Unlock()
}

func (sc *syncTileCache) Close() error {
	close(sc.closec)
	return nil
}

func getCachedOrFetchTile(zoom, x, y int) Tile {

}
