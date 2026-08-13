package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wgt-system/conveyance/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx); err != nil {
		log.Printf("conveyance stopped unexpectedly: %v", err)
		os.Exit(1)
	}
}
