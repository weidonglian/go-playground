package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("require more arguments")
		return
	}
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
	processFile(file)
	fmt.Println("\nTime taken - ", time.Since(s))
}

const kChunkSize = 64 * 1024
const kNumParallel = 32

func processFile(f io.Reader) error {
	chChunks := make(chan []byte, kNumParallel)
	chWalkError := make(chan error, 1)
	chLines := make(chan int, kNumParallel)

	chunkPool := sync.Pool{New: func() interface{} {
		chunk := make([]byte, kChunkSize)
		return chunk
	}}

	go func() {
		defer close(chChunks)
		defer close(chWalkError)
		walkTrunks(f, &chunkPool, chChunks, chWalkError)
	}()

	var wg sync.WaitGroup
	wg.Add(kNumParallel)
	for i := 0; i < kNumParallel; i++ {
		go func() {
			defer wg.Done()
			processChunk(&chunkPool, chChunks, chLines)
		}()
	}

	go func() {
		defer close(chLines)
		wg.Wait()
	}()

	var totalLines int = 0
	for cnt := range chLines {
		totalLines += cnt
	}
	if err := <-chWalkError; err != nil {
		return err
	}
	fmt.Println("total lines:", totalLines)
	return nil
}

func walkTrunks(f io.Reader, chunkPool *sync.Pool, chChunks chan<- []byte, chError chan<- error) {
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
			chError <- err
			return
		}

		nextUntillNewline, err := r.ReadBytes('\n')

		if err != nil && err != io.EOF {
			chError <- err
			return
		}

		if err == nil {
			buf = append(buf, nextUntillNewline...)
		}

		chChunks <- buf
	}
}

func processChunk(chunkPool *sync.Pool, chChunks <-chan []byte, chLines chan<- int) {
	var cnt int = 0
	for chunk := range chChunks {
		logs := bytes.Split(chunk, []byte("\n"))
		// fmt.Println("lines in chunk", len(logs))
		cnt += len(logs)
		/*for _, log := range logs {
			logSlice := bytes.Split(log, []byte(","))
		}*/
		chunkPool.Put(chunk[:cap(chunk)])
	}
	chLines <- cnt
}
