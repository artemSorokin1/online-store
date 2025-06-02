package api

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/artemSorokin1/products-grpc-api/gen/go/product"
)

// server реализует сгенерированный интерфейс ProductsServiceServer.
type server struct {
	pb.UnimplementedProductsServiceServer
	// сюда можно поместить, например, подключение к вашей БД
}

// NewServer возвращает экземпляр нашего сервера.
func NewServer() *server {
	return &server{}
}

// GetProduct — реализация RPC GetProduct(GetProductRequest) → (GetProductResponse)
func (s *server) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	log.Printf("ProductsService.GetProduct: запрос id=%s", req.Id)

	info, _ := structpb.NewStruct(map[string]interface{}{
		"color": "red",
		"size":  "L",
	})

	product := &pb.Product{
		Id:          req.Id,
		Name:        "Тестовый товар",
		Price:       1234,
		ImageUrl:    "https://example.com/img.png",
		Description: "Описание тестового товара",
		Info:        info,
		CreatedAt:   timestamppb.Now(),
		UpdatedAt:   timestamppb.Now(),
		SellerId:    "123e4567-e89b-12d3-a456-426614174001",
		Comments:    []string{"gg", "ff", "hh"},
		Tags:        []string{"тест", "товар", "продукт"},
		Rating:      4.5,
	}

	return &pb.GetProductResponse{
		Product: product,
	}, nil
}

func RunProductsServer(listenAddr string) error {
	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	pb.RegisterProductsServiceServer(grpcServer, NewServer())

	log.Printf("ProductsService gRPC запущен на %s", listenAddr)
	return grpcServer.Serve(lis)
}
