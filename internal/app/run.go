package app

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/wgt-system/conveyance/internal/currentobject"
	"github.com/wgt-system/conveyance/internal/httpapi"
	"github.com/wgt-system/conveyance/internal/persistence/sqlite"
)

type config struct {
	databasePath string
	listenAddr   string
	operation    currentobject.OperationContext
	maxPayload   int
	ready        chan<- net.Addr
}

func Run(ctx context.Context) error {
	return run(ctx, config{
		databasePath: "conveyance.db",
		listenAddr:   "127.0.0.1:8080",
		operation:    currentobject.OperationContext{CanRead: true, CanPublish: true},
		maxPayload:   currentobject.DefaultMaxPayloadSize,
	})
}

func run(ctx context.Context, settings config) error {
	repository, err := sqlite.Open(settings.databasePath)
	if err != nil {
		return err
	}
	defer func() { _ = repository.Close() }()

	service := currentobject.NewService(repository, settings.maxPayload)
	server := &http.Server{Handler: httpapi.NewHandler(service, settings.operation)}
	listener, err := net.Listen("tcp", settings.listenAddr)
	if err != nil {
		return err
	}
	defer func() { _ = listener.Close() }()
	if settings.ready != nil {
		settings.ready <- listener.Addr()
	}

	serveResult := make(chan error, 1)
	go func() { serveResult <- server.Serve(listener) }()

	select {
	case <-ctx.Done():
		shutdownErr := server.Shutdown(context.Background())
		serveErr := <-serveResult
		if shutdownErr != nil {
			return shutdownErr
		}
		if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			return serveErr
		}
		return nil
	case serveErr := <-serveResult:
		if errors.Is(serveErr, http.ErrServerClosed) {
			return nil
		}
		return serveErr
	}
}
