package auth

import (
	"context"
	"os"
	"time"

	grpcauth "github.com/artemSorokin1/Auth-proto/protos/gen/protos/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCAuthClient struct {
	Api    grpcauth.AuthServiceClient
	logger *zap.Logger
}

func New(ctx context.Context,
	logger *zap.Logger,
	timeout time.Duration,
	retriesCount int) *GRPCAuthClient {

	grpcAddres := os.Getenv("GRPC_AUTH_ADDRESS")
	cc, err := grpc.NewClient(grpcAddres, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("failed to create grpc client", zap.Error(err))
		return nil
	}

	return &GRPCAuthClient{
		Api:    grpcauth.NewAuthServiceClient(cc),
		logger: logger,
	}

}
