package md5

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

func stageSourceFileWalker(ctx context.Context, root string) (<-chan string, <-chan error, error) {
	if root == "" {
		return nil, nil, errors.New("empty root is not allowed")
	}

	pathc := make(chan string)
	errc := make(chan error, 1)
	go func() {
		defer close(pathc)
		defer close(errc)
		errc <- filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
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
			case pathc <- path:
			case <-ctx.Done():
				return errors.New("walk canceled")
			}

			return nil
		})
	}()
	return pathc, errc, nil
}

func stageFTFileDigest(ctx context.Context, pathc <-chan string) (<-chan Md5Result, <-chan error, error) {
	const kNumDigesters = 20
	resultc := make(chan Md5Result)
	errc := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(kNumDigesters)
	for i := 0; i < kNumDigesters; i++ {
		go func() {
			defer wg.Done()
			for path := range pathc {
				fmt.Println("start reading file ", path)
				start := time.Now()
				data, err := ioutil.ReadFile(path)
				fmt.Println("finish reading file takes ", path, ", ", time.Since(start))
				select {
				case resultc <- Md5Result{
					path: path,
					sum:  md5.Sum(data),
					err:  err,
				}:
				case <-ctx.Done():
					return
				}
				fmt.Println("checksum takes ", path, ", ", time.Since(start))
			}
		}()
	}

	go func() {
		defer close(resultc)
		defer close(errc)
		wg.Wait()
	}()

	return resultc, errc, nil
}

func stageSink(ctx context.Context, resultc <-chan Md5Result) (map[string]Md5Sum, <-chan error, error) {
	errc := make(chan error, 1)
	m := make(map[string]Md5Sum)
	go func() {
		defer close(errc)
		for result := range resultc {
			if result.err != nil {
				errc <- result.err
				return
			}
			m[result.path] = result.sum
		}
	}()
	return m, errc, nil
}

func WaitForErrors(errcs ...<-chan error) error {
	errc := MergeErrors(errcs...)
	for err := range errc {
		if err != nil {
			return err
		}
	}
	return nil
}

func MergeErrors(errcs ...<-chan error) <-chan error {
	errc := make(chan error, len(errcs))

	var wg sync.WaitGroup
	output := func(ec <-chan error) {
		defer wg.Done()
		for e := range ec {
			errc <- e
		}
	}

	wg.Add(len(errcs))
	for _, c := range errcs {
		go output(c)
	}

	go func() {
		defer close(errc)
		wg.Wait()
	}()

	return errc
}

func Md5AllPar(ctx context.Context, root string) (map[string]Md5Sum, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var errcList []<-chan error
	pathc, errc, err := stageSourceFileWalker(ctx, root)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	resultc, errc, err := stageFTFileDigest(ctx, pathc)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	m, errc, err := stageSink(ctx, resultc)
	if err != nil {
		return nil, err
	}
	errcList = append(errcList, errc)

	if err := WaitForErrors(errcList...); err != nil {
		return nil, err
	}
	return m, nil
}
