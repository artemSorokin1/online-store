package api

import (
	"auth/internal/config"
	"auth/internal/repository/postgres"
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/artemSorokin1/sellers-grpc-api/gen/go/seller"
	"github.com/google/uuid"
)

type server struct {
	pb.UnimplementedSellersServiceServer
	storage *postgres.Storage
}

func NewServer(cfg *config.Config) *server {
	stor, err := postgres.New(cfg)
	if err != nil {
		log.Fatalf("failed to create storage: %v", err)
	}

	return &server{
		storage: stor,
	}
}

func (s *server) GetSeller(ctx context.Context, req *pb.GetSellerRequest) (*pb.GetSellerResponse, error) {
	log.Printf("SellersService.GetSeller: запрос id=%s", req.Id)
	if req.Id == "" {
		return nil, fmt.Errorf("product id is empty")
	}

	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, fmt.Errorf("invalid product id: %w", err)
	}

	grpcSeller, err := s.storage.GetSellerByID(id)
	if err != nil {
		return nil, err
	}

	createdAtTime, err := time.Parse(time.RFC3339, grpcSeller.CreatedAt)
	if err != nil {
		createdAtTime = time.Now()
	}

	seller := &pb.Seller{
		Id:        req.Id,
		Phone:     grpcSeller.Phone,
		Email:     grpcSeller.Email,
		Fullname:  grpcSeller.FullName,
		CreatedAt: timestamppb.New(createdAtTime),
	}

	return &pb.GetSellerResponse{
		Seller: seller,
	}, nil
}

func RunSellersServer(cfg *config.Config) {
	lis, err := net.Listen("tcp", cfg.GrpcCfg.Address)
	if err != nil {
		panic(fmt.Sprintf("failed to listen: %v", err))
	}
	grpcServer := grpc.NewServer()
	pb.RegisterSellersServiceServer(grpcServer, NewServer(cfg))

	log.Printf("SellersService gRPC запущен на %s", cfg.GrpcCfg.Address)
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	log.Println("SellersService gRPC server stopped")
}
