package grpcapp

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	dbgrpc "github.com/db_service/internal/grpc/database"
	"github.com/db_service/internal/service"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type App struct {
	gRPCServer *grpc.Server
	log        *slog.Logger
	port       int
}

func New(log *slog.Logger, getService service.RecordGetter, port int) *App {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived,
			logging.PayloadSent,
		),
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			log.Error("recovered from panic", slog.Any("panic", p))

			return status.Error(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(InterceptorLogger(log), loggingOpts...),
	))

	dbgrpc.Register(gRPCServer, getService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}

}

// InterceptorLogger adapts slog logger to interceptor logger.
func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "app.Run"

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	a.log.Info("started grpc server on ", slog.String("address", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "app.Stop"

	a.log.Info("stopping grpc server", slog.Int("port", a.port), slog.String("op", op))

	a.gRPCServer.GracefulStop()

}
