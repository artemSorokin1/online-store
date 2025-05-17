package grpc

import (
	"auth_service/internal/config"
	"auth_service/internal/repositiry/storage"
	"auth_service/pkg/storage/inmem"
	"fmt"
	api "github.com/artemSorokin1/Auth-proto/protos/gen/protos/proto"
	"google.golang.org/grpc"
	"log/slog"
	"net"
)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

// New создает связб между grpc сервером и реализацией его методов
func New(config config.ServerConfig, s *storage.Storage) *Server {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", config.GRPCPort))
	if err != nil {
		panic(err)
	}
	fmt.Println(lis.Addr().String())
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)

	refreshStor := inmem.NewRedisStorage()

	api.RegisterAuthServiceServer(grpcServer, &AuthService{stor: s, refreshInMemStorage: refreshStor})

	return &Server{
		grpcServer,
		lis,
	}
}

func (s *Server) MustStart() {
	slog.Info("grpc auth server start")
	err := s.grpcServer.Serve(s.listener)
	if err != nil {
		panic(err)
	}
}

func (s *Server) GracefulStop() {
	slog.Info("stopping grpc auth server")
	s.grpcServer.GracefulStop()
	slog.Info("grpc auth server stopped")
}
