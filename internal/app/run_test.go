package app

import (
	"context"
	"net"
	"path/filepath"
	"sync"
	"testing"

	"github.com/wgt-system/conveyance/internal/currentobject"
)

type observedContext struct {
	context.Context
	observed chan struct{}
	once     sync.Once
}

func TestRunWithEphemeralLoopbackPortStopsOnCancellation(t *testing.T) {
	baseCtx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	ready := make(chan net.Addr, 1)
	go func() {
		result <- run(baseCtx, config{
			databasePath: filepath.Join(t.TempDir(), "conveyance.db"),
			listenAddr:   "127.0.0.1:0",
			operation:    currentobject.OperationContext{CanRead: true, CanPublish: true},
			maxPayload:   currentobject.DefaultMaxPayloadSize,
			ready:        ready,
		})
	}()
	if addr := <-ready; addr.String() == "" {
		t.Fatal("server reported an empty listen address")
	}
	cancel()
	if err := <-result; err != nil {
		t.Fatalf("run returned an error after cancellation: %v", err)
	}
}

func (ctx *observedContext) Done() <-chan struct{} {
	ctx.once.Do(func() {
		close(ctx.observed)
	})
	return ctx.Context.Done()
}

func TestRunWaitsForCancellation(t *testing.T) {
	baseCtx, cancel := context.WithCancel(context.Background())
	ctx := &observedContext{
		Context:  baseCtx,
		observed: make(chan struct{}),
	}
	result := make(chan error, 1)

	go func() {
		result <- run(ctx, config{
			databasePath: filepath.Join(t.TempDir(), "conveyance.db"),
			listenAddr:   "127.0.0.1:0",
			operation:    currentobject.OperationContext{CanRead: true, CanPublish: true},
			maxPayload:   currentobject.DefaultMaxPayloadSize,
		})
	}()

	<-ctx.observed
	select {
	case err := <-result:
		t.Fatalf("Run returned before cancellation: %v", err)
	default:
	}

	cancel()

	if err := <-result; err != nil {
		t.Fatalf("Run returned an error after cancellation: %v", err)
	}
}
