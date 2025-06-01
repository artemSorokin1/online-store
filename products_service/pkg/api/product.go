package api

import (
	"context"
	"log"
	"net"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/artemSorokin1/products-grpc-api/gen/go/product"
	"github.com/artemSorokin1/products-grpc-api/internal/repository"
)

// server реализует сгенерированный интерфейс ProductsServiceServer.
type server struct {
	pb.UnimplementedProductsServiceServer
	repo *repository.ProductRepository
}

// NewServer возвращает экземпляр нашего сервера.
func NewServer(repo *repository.ProductRepository) *server {
	return &server{repo: repo}
}

// GetProduct — реализация RPC GetProduct(GetProductRequest) → (GetProductResponse)
func (s *server) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	log.Printf("ProductsService.GetProduct: запрос id=%s", req.Id)

	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, err
	}

	product, err := s.repo.GetProduct(ctx, id)
	if err != nil {
		return nil, err
	}

	if product == nil {
		return nil, nil
	}

	info, err := structpb.NewStruct(map[string]interface{}{
		"info": string(product.Info),
	})
	if err != nil {
		return nil, err
	}

	return &pb.GetProductResponse{
		Product: &pb.Product{
			Id:          strconv.FormatInt(product.ID, 10),
			Name:        product.Name,
			Price:       product.Price,
			ImageUrl:    product.ImageURL.String,
			Description: product.Description.String,
			Info:        info,
			CreatedAt:   timestamppb.New(product.CreatedAt),
			UpdatedAt:   timestamppb.New(product.UpdatedAt),
			SellerId:    product.SellerID.String(),
			Comments:    product.Comments,
			Rating:      product.Rating,
		},
	}, nil
}

func RunProductsServer(listenAddr string, repo *repository.ProductRepository) error {
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterProductsServiceServer(grpcServer, NewServer(repo))

	log.Printf("ProductsService gRPC запущен на %s", listenAddr)
	return grpcServer.Serve(lis)
}
