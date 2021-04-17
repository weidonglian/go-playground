package impl

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"sync"
	"time"
)

func fileWalker(ctx context.Context, root string) (<-chan string, <-chan error) {
	chPaths := make(chan string, 20)
	chError := make(chan error, 1)
	go func() {
		defer close(chPaths)
		defer close(chError)
		chError <- filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			info, err := d.Info()
			if err != nil {
				return err
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			select {
			case chPaths <- path:
			case <-ctx.Done():
				return errors.New("walk canceled")
			}

			return nil
		})
	}()
	return chPaths, chError
}

func fileDigester(ctx context.Context, paths <-chan string, results chan<- Md5Result) {
	for path := range paths {
		fmt.Println("start reading file ", path)
		start := time.Now()
		data, err := ioutil.ReadFile(path)
		fmt.Println("finish reading file takes ", path, ", ", time.Since(start))
		select {
		case <-ctx.Done():
			return
		case results <- Md5Result{
			path: path,
			sum:  md5.Sum(data),
			err:  err,
		}:
		}
		fmt.Println("checksum takes ", path, ", ", time.Since(start))
	}
}

func Md5AllPar(ctx context.Context, root string) (map[string]Md5Sum, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	chPaths, chError := fileWalker(ctx, root)

	const numDigesters = 20
	chResults := make(chan Md5Result)
	var wg sync.WaitGroup
	wg.Add(numDigesters)
	for i := 0; i < numDigesters; i++ {
		go func() {
			fileDigester(ctx, chPaths, chResults)
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(chResults)
	}()

	m := make(map[string]Md5Sum)
	for result := range chResults {
		if result.err != nil {
			return nil, result.err
		}
		m[result.path] = result.sum
	}
	if err := <-chError; err != nil {
		return nil, err
	}
	return m, nil
}
