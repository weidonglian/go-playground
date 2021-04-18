package main

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/weidonglian/go-playground/util"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("cmd argument should be 'seq <path> or par <path>'")
		return
	}

	if os.Args[1] != "seq" && os.Args[1] != "par" {
		fmt.Println("cmd argument should be 'seq <path> or par <path>'")
		return
	}

	chanTerm := make(chan os.Signal)
	signal.Notify(chanTerm, syscall.SIGTERM, syscall.SIGINT)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() { // add a hookup to cancel only when term or ctrl+c is triggered
		<-chanTerm   // block here until notified with ctrl+c or killed
		cancelFunc() // tell other context we are canceling please exit as soon as possible
	}()

	var m map[string]Md5Sum
	var err error
	if os.Args[1] == "par" {
		m, err = Md5AllPar(ctx, os.Args[2])
	} else {
		m, err = Md5AllSeq(ctx, os.Args[2])
	}

	if err != nil {
		fmt.Println("Failed Md5AllSeq with err:", err)
		return
	}
	paths := make([]string, 0, len(m))
	for path := range m {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	// for _, path := range paths {
	// fmt.Printf("%x  %s\n", m[path], path)
	// }
	fmt.Println("gracefully shutdown:)")
}

type Md5Sum [md5.Size]byte

type Md5Result struct {
	path string
	sum  Md5Sum
	err  error
}

func Md5AllSeq(ctx context.Context, root string) (map[string]Md5Sum, error) {
	m := make(map[string]Md5Sum)
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if ctx.Err() == context.Canceled {
			return context.Canceled
		}

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

		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		m[path] = md5.Sum(data)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return m, nil
}

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

	if err := util.WaitForPipeline(errcList...); err != nil {
		return nil, err
	}
	return m, nil
}
