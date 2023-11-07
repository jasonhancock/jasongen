package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	ver "github.com/jasonhancock/cobra-version"
	"github.com/jasonhancock/jasongen/cmd"
)

// These variables are populated by goreleaser when the binary is built.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	go func() {
		defer close(done)
		cmd.Execute(ctx, &wg, ver.Info{
			Version: version,
			Commit:  commit,
			Date:    date,
			Go:      runtime.Version(),
		})
	}()

	<-done
	cancel()
	wg.Wait()
}
