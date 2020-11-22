package helpers

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func CancelRoot () {
	rootContext := context.Background()
	rootContext, cancelRoot := context.WithCancel(rootContext)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		log.Info("Received signal %s", sig)
		cancelRoot()
	}()

	<-rootContext.Done()
	log.Info("Shut down daemon.")
}