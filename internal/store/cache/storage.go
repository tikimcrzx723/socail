package cache

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/tikimcrzx723/social/internal/store"
)

type Storage struct {
	Users interface {
		Get(ctx context.Context, userID int64) (*store.User, error)
		Set(ctx context.Context, user *store.User) error
		Delete(ctx context.Context, userID int64)
	}
}

func NewRedisStorage(rdb *redis.Client) Storage {
	return Storage{
		Users: &UserStore{rdb: rdb},
	}
}
