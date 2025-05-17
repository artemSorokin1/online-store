package inmem

import "context"

type RefreshTokenStorage interface {
	SaveToken(ctx context.Context, userId int64, token string) error
	GetToken(ctx context.Context, userId int64) (string, error)
	RemoveToken(ctx context.Context, userId int64) error
}
