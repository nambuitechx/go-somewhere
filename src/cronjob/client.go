package main

import (
    "context"
    "time"

    r2 "github.com/redis/go-redis/v9"
)

// Config holds Redis connection options
type Config struct {
    Addr string
    Password string
    DB int
}

// Client wraps redis.Client
type Client struct {
    *r2.Client
}

func NewClient(cfg Config) *Client {
    r := r2.NewClient(&r2.Options{
        Addr: cfg.Addr,
        Password: cfg.Password,
        DB: cfg.DB,
    })

    // Ping
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    _ = r.Ping(ctx).Err()

    return &Client{r}
}
