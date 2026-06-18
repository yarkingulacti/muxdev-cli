package platform

import (
	"context"
	"os"
	"os/signal"
)

func HandleInterrupt(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, interruptSignals()...)
	defer signal.Stop(sigCh)

	<-sigCh
	cancel()
}
