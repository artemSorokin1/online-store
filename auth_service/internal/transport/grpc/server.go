package grpc

import (
	"auth_service/internal/config"
	"auth_service/internal/repositiry/storage"
	"auth_service/pkg/storage/inmem"
	"fmt"
	"log/slog"
	"net"

	api "github.com/artemSorokin1/Auth-proto/protos/gen/protos/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	logger     *zap.Logger
}

// New создает связб между grpc сервером и реализацией его методов
func New(config config.ServerConfig, s *storage.Storage, logger *zap.Logger) *Server {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", config.GRPCPort))
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)

	refreshStor := inmem.NewRedisStorage()

	api.RegisterAuthServiceServer(grpcServer, &AuthService{stor: s, refreshInMemStorage: refreshStor})

	return &Server{
		grpcServer,
		lis,
		logger,
	}
}

func (s *Server) MustStart() {
	slog.Info("grpc auth server start")
	err := s.grpcServer.Serve(s.listener)
	if err != nil {
		s.logger.Fatal("failed to start grpc server", zap.Error(err))
	}
}

func (s *Server) GracefulStop() {
	s.logger.Info("grpc auth server stopping")
	s.grpcServer.GracefulStop()
	s.logger.Info("grpc auth server stopped")
}
