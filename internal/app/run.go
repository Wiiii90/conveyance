package app

import "context"

func Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
