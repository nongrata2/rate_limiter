package repositories

import (
	"context"
	"time"
)

type Client struct {
	Key        string        `json:"key"`
	Capacity   int64         `json:"capacity"`
	RefillRate time.Duration `json:"refill_rate"`
	Unlimited  bool          `json:"unlimited"`
	CreatedAt  time.Time     `json:"created_at"`
}

type DBInterface interface {
	AddClient(ctx context.Context, client Client) error
	GetClient(ctx context.Context, key string) (Client, error)
	ListClients(ctx context.Context) ([]Client, error)
	DeleteClient(ctx context.Context, key string) error
}
