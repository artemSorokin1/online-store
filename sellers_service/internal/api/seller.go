package api

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	// импорт сгенерированного SellersService
	pb "github.com/artemSorokin1/sellers-grpc-api/gen/go/seller"
)

// server реализует sгенерированный интерфейс SellersServiceServer.
type server struct {
	pb.UnimplementedSellersServiceServer
	// сюда можно положить соединение с БД
}

// NewServer возвращает экземпляр сервера Sellers.
func NewServer() *server {
	return &server{}
}

// GetSeller — реализация RPC GetSeller(GetSellerRequest) → (GetSellerResponse)
func (s *server) GetSeller(ctx context.Context, req *pb.GetSellerRequest) (*pb.GetSellerResponse, error) {
	log.Printf("SellersService.GetSeller: запрос id=%s", req.Id)

	// Здесь вместо заглушки должен быть запрос к БД по req.Id
	seller := &pb.Seller{
		Id:        req.Id,
		Phone:     "+7 123 456-78-90",
		Email:     "seller@example.com",
		Fullname:  "Иван Петров",
		CreatedAt: timestamppb.Now(),
	}

	return &pb.GetSellerResponse{
		Seller: seller,
	}, nil
}

// Run запускает gRPC-сервер Sellers на указанном адресе (например, ":50052").
func RunSellersServer(listenAddr string) error {
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterSellersServiceServer(grpcServer, NewServer())

	log.Printf("SellersService gRPC запущен на %s", listenAddr)
	return grpcServer.Serve(lis)
}
