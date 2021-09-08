# go-playground

This is my `golang` learning note.

There is also another learning note about the back-end stack: `microservce`,
`messaging queu`, `http`, `REST` and `GraphQL` API, `authentication`, `deployment`. see
this [repo](https://github.com/weidonglian/notes-app) for more details.

## Channels

channel is normally used to communicate between go routines. There are two types of channels: `buffered`

and `unbuffered` .
`buffered` will only block when sending to a full-buffered or receiving from an empty, `unbuffered` will always block
when sending or receiving.

* declare a channel: `var ch chan int`
* declare a read-only (receive from) channel: `var ch <-chan int` arrow is on the left of `chan`
* declare a write-only (send to) channel: `var ch chan<- int` arrow is on the right of `chan`
* create a unbuffered channel: `ch := make(chan int)` or `var ch chan int; ch = make(chan int)`
* create a buffered channel with the given size `sz`: `ch := make(chan int, sz)`

we can perform read (receive) operation or write (send) operation on the channel from different go routines. `chan`

object is thread-safe when operating from different goroutines.

A single send or receive operation:

* read (receive) operation: `val := <-ch`
* write (send) operation: `ch <- val`

With `for-range` to perform receive operation from channel, we have to make sure the channel is closed when the
sender/provider has sent out all the values via the channel; otherwise the `for-range` receive will block forever.
`for v := range ch { fmt.println("value is:%d", v) }`

A design consideration about if we should close a channel or leave the channel open forever? An authoritative answer
from the Google Group is that we do not need to explicitly close a channel unless the receiver side relies on the close
signal, i.e. `val, ok := <-ch` if `ok` is `false` then it indicates the channel has closed. Also, `for-range` relies on
a closed channel to terminate the corresponding goroutine; otherwise, it will block there forever.

Inside one goroutine, how to both send or receive from one or multiple channels? The `select` or `for-select` to rescue.

``` go
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

Which side is responsible or takes the ownership of the channel? Conceptually, it should be the `sender` side of the
channel that should close it if needed. A conventional pattern is that

* we initiate a new channel
* pass it to the `sender` (i.e. producer in another goroutine), when all values have been processed (i.e. waitgroup),
  close it.
* pass it to the `receiver` (i.e. consumer in another goroutine), receive values from channel and also check if channel

  has been closed. It is the sender side that should close the channel when all values have been processed.

## Pipeline

A pipeline is a data processing pipe includes `source` -> `filter` + `transform` +... -> `sink` states. A common pattern
is defined as follows:

* source stage

`func sourceStage(ctx Context, input PipelineInput)(output <-chan OutputFromSource, errc <-chan error, err error)` .

* filter/transform stage

`function filter/transformStage(ctx Context, input <-chan OutputFromSource)(output <-chan OutputFromFT, errc <-chan error, err error)`

this stage can include many stages for a complex pipeline. It could fan-out or fan-in by your design.

* sink stage

`func sinkStage(ctx Context, input <-chan OutputFromFT, finalOutput PipelineOutput)(errc <-chan error, err error)` .

It is a good practice to always return `errc <-chan error, err error` pair for every state. The `err` is returned when
the corresponding stage has not started, we can perform an early fail and return. `errc` is used to indicate and tells
the `pipeline` the error occurs inside the goroutine of the corresponding state.

In the pipeline, we could merge all the stage `errc` into a dedicated `error` channel (fan-in all stage errors), we
could then `for-range` to wait for any stage error.

below is a simple example of the pipeline [demo](https://medium.com/statuscode/pipeline-patterns-in-go-a37bb3a7e61d)
which is a clean and clear pattern to handle the pipeline:

``` go
func runSimplePipeline(ctx context.Context, input PipelineInput, output PipelineOutput) error {
 ctx, cancelFunc := context.WithCancel(ctx)
 defer cancelFunc()
 var errcList []<-chan error
 // Source pipeline stage.
 sourceOutput, errc, err := sourceStage(ctx, input)
 if err != nil {
  return err
 }
 errcList = append(errcList, errc)
 // Filter/Transformer pipeline stage.
 ftOutput, errc, err := filter_or_transformStage(ctx, sourceOutput)
 if err != nil {
  return err
 }
 errcList = append(errcList, errc)

 // Might have other Filter/Transformer
 // ...

 // Sink pipeline stage.
 errc, err = sinkStage(ctx, ftOutput, output)
 if err != nil {
  return err
 }
 errcList = append(errcList, errc)
 fmt.Println("Pipeline started. Waiting for pipeline to complete.")
 return WaitForPipeline(errcList...)
}

// WaitForPipeline waits for results from all error channels.
// It returns early on the first error.
func WaitForPipeline(errs ...<-chan error) error {
 errc := MergeErrors(errs...)
 for err := range errc {
  if err != nil {
   return err
  }
 }
 return nil
}

// MergeErrors merges multiple channels of errors.
// Based on https://blog.golang.org/pipelines.
func MergeErrors(cs ...<-chan error) <-chan error {
 var wg sync.WaitGroup
 // We must ensure that the output channel has the capacity to
 // hold as many errors
 // as there are error channels.
 // This will ensure that it never blocks, even
 // if WaitForPipeline returns early.
 out := make(chan error, len(cs))
 // Start an output goroutine for each input channel in cs.  output
 // copies values from c to out until c is closed, then calls
 // wg.Done.
 output := func(c <-chan error) {
  for n := range c {
   out <- n
  }
  wg.Done()
 }
 wg.Add(len(cs))
 for _, c := range cs {
  go output(c)
 }
 // Start a goroutine to close out once all the output goroutines
 // are done.  This must start after the wg.Add call.
 go func() {
  wg.Wait()
  close(out)
 }()
 return out
}
```

Another challenge in the pipeline handling is how to keep the output order the same as input order? This is not easy to
done in an elegant way, we need to keep the jobs order into a `Queue` before sending to the job processer, then before
sink any job, we need to check if the job matches the queue's current head, if not we need to push the processed job
into an order list.

## Context

A `context.Context` is a useful and elegant helper to communicate between pipeline stages (i.e. multiple goroutines) or
even a single goroutine.

The context API is thread-safe (i.e. can be called among multiple goroutines).

There are two `demonon` contexts which are the root paraents which is not cancellable, i.e. `context.Background()`
and `context.Todo()`.

How to use it?

```go
ctx, cancelFunc := context.WithCancel(context.Background())
defer cancelFunc()
```

`context.Deadline` and `context.Timeout` are just two variants of `context.WithCancel`.

We could check the state of context in two ways:

* via error `ctx.Error() == context.ErrorCanceled` to indicate the cancellation of context.
* via channel `<-ctx.Done()` this channel will be unblocked when cancel is triggered. This `done` channel is useful when
  combined with other send or receive channels together with `select-case` to unblock the `goroutine` and cleanup the
  goroutines and channels resources.

Note the `context` is derivable, when the parent context is canceled, all the derived contexts will be canceled as well.

How to gracefully shut down an app using context?

```go
func main() {
  chanTerm := make(chan os.Signal)
  signal.Notify(chanTerm, os.syscall.SIGTERM, os.syscall.SIGINIT)
  ctx, cancelFunc := context.WithCancel(context.Background())
  go func() {
    // put the hooking the Ctrl+C or termination into a goroutine will not block the normal app exit
    <-chanTerm
    cancelFunc()
  }()

  // ...
  // normal app code comes here
  // ...
}
```

Another aspect of `context.Context` is to transfer `key-value` pair between API and package boundaries. This should only
be used to pass argument between API boundaries instead of passing normal optional arguments. In the `http`
handler, `context.WithValue` is intensively used to pass information between different `http.HttpHandler` chains.

## Buffer, Bytes, IoBuffer and FileHandling

There are tons of APIs to handle a small file. The challenge is to handle the huge file in a CPU efficient and memory
optimized way, e.g. a file around 100 GB or 1 TB size.

Possible considerations:

* Read line by line or scan line by line is memory optimal but takes quite long time due to the file context switches.
* Read the whole into memory buffer and then process it is CPU efficient, but mempry hungry (i.e. out of memory).

Above are two extreme cases. It does not really matter for smaller files.

A middle ground solution would be read the file chunk by chunk to leverage the memory and cpu efficiency:

```go
const kBufferSize = 64 * 1024 // 64k
chunkPool := sync.Pool{New: func() { return make([]byte, kBufferSize)}}
file := os.open(filename)
reader := bufio.NewReaderSize(file, kBufferSize)
buf := chunkPool.Get()
n, err := reader.Read()
// check n and err to decide if we should continue
chunkPool.Put(buf[:cap(buf)])
```

The `sync.Pool` is important to avoid the frequent memory allocation that japatize the performance of gabbage collector.
With the help `pool` we will allocate less frequenctly.

Another keep point is to avoid convert the buffer from `[]byte` into `string`. `bytes` packages in golang has all kinds
of utilities similar to string to handle directly on `[]byte`. This avoids the memory allocation a lot.

The key focus should always on the actual data structure, like in C++:).

## Synchronization

### sync.pool

sync.pool is used to reuse a temporary object via pool. We can put and get a random object from the pool. This is very
useful if we want to reuse allocated buffers to avoid the burden of memory allocation.

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
    go run xxxx.go
```

* Build and check if there is any compilation error

```sh
    go build
```

* Install into the bin folder of the $GOPATH

```sh
  go install
```

* Run the unit test in each package TestXXX in xxx_test.go

```sh
  go test ./...
```

* Run the benchmark in each package BenchYYY in yyy_test.go

```sh
  go test -bench=. ./... 
```
