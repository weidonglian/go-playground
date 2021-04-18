package main

import (
	"context"
	"fmt"
	"github.com/weidonglian/go-playground/md5"
	"os"
	"os/signal"
	"sort"
	"syscall"
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

	var m map[string]md5.Md5Sum
	var err error
	if os.Args[1] == "par" {
		m, err = md5.Md5AllPar(ctx, os.Args[2])
	} else {
		m, err = md5.Md5AllSeq(ctx, os.Args[2])
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
