package util

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func HandleSignalShutdown(cancelRoot context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		log.Info("Received signal %s", sig)
		cancelRoot()
	}()
}
