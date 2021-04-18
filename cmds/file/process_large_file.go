package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/weidonglian/go-playground/util"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("require more arguments")
		return
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	termc := make(chan os.Signal)
	signal.Notify(termc, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-termc
		cancelFunc()
	}()

	fileName := os.Args[1]
	file, err := os.Open(fileName)

	if err != nil {
		fmt.Println("cannot able to read the file", err)
		return
	}

	defer file.Close()

	filestat, err := file.Stat()
	if err != nil {
		fmt.Println("Could not able to get the file stat")
		return
	}

	fileSize := filestat.Size()
	fmt.Println("file size is ", fileSize)
	s := time.Now()
	if err := runProcessFilePipeline(ctx, file); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Time taken - \n", time.Since(s))
}

const kChunkSize = 1024 * 1024
const kNumParallel = 32

// This pipeline will count the number of lines and it is easy to extend to parse the information line by line.
// It reads trunks of data and then try to continue read to next line's boundary, then chunks will devide the whole
// file across the line boundary.
func runProcessFilePipeline(ctx context.Context, f io.Reader) error {
	ctx, cancelFunc := context.WithCancel(ctx)
	defer cancelFunc()

	chunkPool := sync.Pool{New: func() interface{} {
		return make([]byte, kChunkSize)
	}}

	var errcList []<-chan error
	chunkc, errc, err := stageSourceWalkTrunks(ctx, f, &chunkPool)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	linesc, errc, err := stageFTProcessChunk(ctx, &chunkPool, chunkc)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	totalLines, errc, err := stageSink(linesc)
	if err != nil {
		return nil
	}
	errcList = append(errcList, errc)
	if err := util.WaitForPipeline(errcList...); err != nil {
		return err
	}
	// Now
	fmt.Println("total lines:", *totalLines)
	return nil
}

func stageSourceWalkTrunks(ctx context.Context, f io.Reader, chunkPool *sync.Pool) (<-chan []byte, <-chan error, error) {
	chunkc := make(chan []byte)
	errc := make(chan error, 1)

	go func() {
		defer close(chunkc)
		defer close(errc)

		r := bufio.NewReaderSize(f, kChunkSize)
		iChunk := 0
		for {
			buf := chunkPool.Get().([]byte)
			n, err := r.Read(buf)
			// fmt.Printf("Chunk %d with len(%d) and cap(%d): %d %v\n", iChunk, len(buf), cap(buf), n, err)
			iChunk++
			buf = buf[:n]

			if err != nil {
				if err == io.EOF {
					break
				}
				errc <- err
				return
			}

			nextUntillNewline, err := r.ReadBytes('\n')

			if err != nil && err != io.EOF {
				errc <- err
				return
			}

			if err == nil {
				buf = append(buf, nextUntillNewline...)
			}

			select {
			case chunkc <- buf:
			case <-ctx.Done():
				errc <- ctx.Err()
				return
			}
		}
	}()

	return chunkc, errc, nil
}

func stageFTProcessChunk(ctx context.Context, chunkPool *sync.Pool, chunkc <-chan []byte) (<-chan int, <-chan error, error) {
	linesc := make(chan int)
	errc := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Add(kNumParallel)
	for i := 0; i < kNumParallel; i++ {
		go func() {
			defer wg.Done()
			for chunk := range chunkc {
				logs := bytes.Split(chunk, []byte("\n"))
				// fmt.Println("lines in chunk", len(logs))
				select {
				case linesc <- len(logs):
				case <-ctx.Done():
					errc <- ctx.Err()
					return
				}
				/*for _, log := range logs {
					logSlice := bytes.Split(log, []byte(","))
				}*/
				chunkPool.Put(chunk[:cap(chunk)])
			}
		}()
	}

	go func() {
		defer close(linesc)
		defer close(errc)
		wg.Wait()
	}()

	return linesc, errc, nil
}

func stageSink(linesc <-chan int) (*int, <-chan error, error) {
	errc := make(chan error)
	totalLines := 0
	go func() {
		defer close(errc)
		for lines := range linesc {
			totalLines += lines
		}
	}()
	return &totalLines, errc, nil
}
