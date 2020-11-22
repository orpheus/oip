package util

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func CancelRoot (ctx context.Context, cancelRoot context.CancelFunc) {

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		log.Info("Received signal %s", sig)
		cancelRoot()
	}()

	<-ctx.Done()
	log.Info("Shut down daemon.")
}