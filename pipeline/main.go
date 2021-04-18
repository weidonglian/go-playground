package main

import (
	"context"
	"fmt"
	"sync"
)

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
