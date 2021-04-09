package channel

import (
	"fmt"
	"sync"
	"testing"
)

import "github.com/stretchr/testify/assert"

type StructID struct {
	id int
}

func TestChannelBufferChanRef(t *testing.T) {
	// a lesson here is that you should not pass a channel in one go routine to another, if you have to then it is better
	// pass the value instead of the channel variable to another go routine.
	channel := make(chan StructID, 5)
	chanResults := make(chan int, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for a := range channel {
			wg.Add(1)
			go func() {
				defer wg.Done()
				t.Log(a.id, ",", fmt.Sprintf("%T\n", a))
				chanResults <- a.id
			}()
		}

	}()

	for i := 0; i < 10; i++ {
		channel <- StructID{id: i}
	}
	close(channel)
	wg.Wait()
	close(chanResults)
	hasDuplication := false
	resultsMap := make(map[int]int)
	for res := range chanResults {
		if _, ok := resultsMap[res]; ok {
			hasDuplication = true
		}
		resultsMap[res] = res
	}
	assert.True(t, hasDuplication)
}

func TestChannelBufferChanValue(t *testing.T) {
	// a lesson here is that you should not pass a channel in one go routine to another, if you have to then it is better
	// pass the value instead of the channel variable to another go routine.
	chanResults := make(chan int, 10)
	channel := make(chan StructID, 5)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for a := range channel {
			wg.Add(1)
			go func(item StructID) {
				defer wg.Done()
				t.Log(item.id, ",", fmt.Sprintf("%T\n", item))
				chanResults <- item.id
			}(a)
		}
	}()

	for i := 0; i < 10; i++ {
		channel <- StructID{id: i}
	}
	close(channel)

	wg.Wait()

	close(chanResults)

	resultsMap := make(map[int]int)
	for res := range chanResults {
		_, ok := resultsMap[res]
		assert.False(t, ok)
		resultsMap[res] = res
	}
}
