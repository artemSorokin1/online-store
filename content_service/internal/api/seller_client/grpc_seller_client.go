package seller_client

import (
	"context"
	"time"

	pb "github.com/artemSorokin1/sellers-grpc-api/gen/go/seller"
	"google.golang.org/grpc"
)

type SellersClient struct {
	grpcClient pb.SellersServiceClient
}

func NewSellersClient() (*SellersClient, error) {
	conn, err := grpc.Dial(":50052", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return &SellersClient{grpcClient: pb.NewSellersServiceClient(conn)}, nil
}

func (c *SellersClient) GetSeller(ctx context.Context, productID string) (*pb.Seller, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	resp, err := c.grpcClient.GetSeller(ctx, &pb.GetSellerRequest{Id: productID})
	if err != nil {
		return nil, err
	}
	return resp.GetSeller(), nil
}
