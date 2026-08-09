package app

import (
	"context"
	"sync"
	"testing"
)

type observedContext struct {
	context.Context
	observed chan struct{}
	once     sync.Once
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
		result <- Run(ctx)
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
