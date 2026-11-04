package storage

import "context"

type Storage struct{}

func New() *Storage {
        return &Storage{}
}

func (s *Storage) Ping(ctx context.Context) error {
        return nil
}