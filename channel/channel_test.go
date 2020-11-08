package channel

import (
	"fmt"
	"sync"
	"testing"
)

type A struct {
	id int
}

func TestChannelBufferChanRef(t *testing.T) {
	// a lesson here is that you should not pass a channel in one go routine to another, if you have to then it is better
	// pass the value instead of the channel variable to another go routine.
	channel := make(chan A, 5)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for a := range channel {
			wg.Add(1)
			go func() {
				defer wg.Done()
				t.Log(a.id, ",", fmt.Sprintf("%T\n", a))
			}()
		}

	}()

	for i := 0; i < 10; i++ {
		channel <- A{id: i}
	}
	close(channel)

	wg.Wait()
}

func TestChannelBufferChanValue(t *testing.T) {
	// a lesson here is that you should not pass a channel in one go routine to another, if you have to then it is better
	// pass the value instead of the channel variable to another go routine.
	channel := make(chan A, 5)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for a := range channel {
			wg.Add(1)
			go func(item A) {
				defer wg.Done()
				t.Log(item.id, ",", fmt.Sprintf("%T\n", item))
			}(a)
		}

	}()

	for i := 0; i < 10; i++ {
		channel <- A{id: i}
	}
	close(channel)

	wg.Wait()
}