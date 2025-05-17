package auth

import (
	"context"
	"dlivery_service/delivery_service/pkg/logger"
	grpcauth "github.com/artemSorokin1/Auth-proto/protos/gen/protos/proto"
	"google.golang.org/grpc"
	"time"
)

const AddressAuthGRPCServer = "grpc_auth_server:8082"

type GRPCAuthClient struct {
	Api grpcauth.AuthServiceClient
	log *logger.Logger
}

func New(ctx context.Context,
	timeout time.Duration,
	retriesCount int) *GRPCAuthClient {
	cc, err := grpc.Dial(AddressAuthGRPCServer, grpc.WithInsecure())
	if err != nil {
		log := logger.GetLoggerFromContext(ctx)
		log.Error(ctx, "failed to create grpc client")
		return nil
	}

	return &GRPCAuthClient{
		Api: grpcauth.NewAuthServiceClient(cc),
	}

}
