package product_client

import (
	"context"
	"time"

	pb "github.com/artemSorokin1/products-grpc-api/gen/go/product"
	"google.golang.org/grpc"
)

// ProductsClient wraps gRPC client for ProductsService.
type ProductsClient struct {
	grpcClient pb.ProductsServiceClient
}

func NewProductsClient() (*ProductsClient, error) {
	conn, err := grpc.Dial("products_service:50051", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &ProductsClient{
		grpcClient: pb.NewProductsServiceClient(conn),
	}, nil
}

func (c *ProductsClient) GetProduct(ctx context.Context, id string) (*pb.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.grpcClient.GetProduct(ctx, &pb.GetProductRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return resp.GetProduct(), nil
}
