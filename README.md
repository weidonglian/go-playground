# go-playground

## Channels

channel is normally used to communicate between go routines. There are two types of channels: `buffered`
and `unbuffered`.
`buffered` will only block when sending to a full-buffered or receiving from an empty, `unbuffered` will always block
when sending or receiving.

- declare a channel: `var ch chan int`
- declare a read-only (receive from) channel: `var ch <-chan int` arrow is on the left of `chan` 
- declare a write-only (send to) channel: `var ch chan<- int` arrow is on the right of `chan`
- create a unbuffered channel: `ch := make(chan int)` or `var ch chan int; ch = make(chan int)`
- create a buffered channel with the given size `sz`: `ch := make(chan int, sz)`

we can perform read (receive) operation or write (send) operation on the channel from different go routines. `chan` 
object is thread-safe when operating from different goroutines. 

A single send or receive operation: 
- read (receive) operation: `val := <-ch`
- write (send) operation: `ch <- val`

With `for-range` to perform receive operation from channel, we have to make sure the channel is closed when the 
sender/provider has sent out all the values via the channel; otherwise the `for-range` receive will block forever.
`for v := range ch { fmt.println("value is:%d", v) }`

A design consideration about if we should close a channel or leave the channel open forever? 
An authoritative answer from the Google Group is that we do not need to explicitly close a channel unless the receiver 
side relies on the close signal, i.e. `val, ok := <-ch` if `ok` is `false` then it indicates the channel has closed. 
Also, `for-range` relies on a closed channel to terminate the corresponding goroutine; otherwise, it will block there 
forever. 

Inside one goroutine, how to both send or receive from one or multiple channels? The `select` or `for-select` to rescue.

```go
func fibonacci(c, quit chan struct{}) {
	x, y := 0, 1
	for {
		select {
		case c <- x:
			x, y = y, x+y
		case <-quit:
			fmt.Println("quit")
			return
		}
	}
}
```

Which side is responsible or takes the ownership of the channel? 
Conceptually, it should be the `sender` side of the channel that should close it if needed. 
A conventional pattern is that 
- we initiate a new channel
- pass it to the `sender` (i.e. producer in another goroutine), when all values have been processed (i.e. waitgroup), close it. 
- pass it to the `receiver` (i.e. consumer in another goroutine), receive values from channel and also check if channel 
  has been closed.  
It is the sender side that should close the channel when all values have been processed.

## Synchronization

### sync.pool

sync.pool is used to reuse a temporary object via pool. We can put and get a random object from the pool.

## Misc

### `struct{}` vs `interface{}`

`struct{}` indicates an empty struct that we know for sure what type it stands for, i.e. empty.
`interface{}` could be used to hold any value, we do not know for sure what it holds. 

In practice, `struct{}` is better than `interface{}` when we want to pass a flag or quit/done channel.


## How to Build & Install

Please read the [Golang](https://golang.org/doc/code.html) for details how to organize the workspace.

Go into a directory of the package, i.e. tree

* Run the file with `func main()`

    ```sh
    $ go run xxxx.go
    ```

* Build and check if there is any compilation error

    ```sh
    $ go build
    ``` 

* Install into the bin folder of the $GOPATH

  ```sh
  $ go install
  ```

* Run the unit test in each package TestXXX in xxx_test.go

  ```sh
  $ go test
  ```