package usecase

import (
        "context"

        "agnis-crm/internal/storage"
)

type Usecase struct {
        storage *storage.Storage
}

func New(st *storage.Storage) *Usecase {
        return &Usecase{storage: st}
}

// Health проверяет, что все нижележащие слои живы.
func (u *Usecase) Health(ctx context.Context) error {
        return u.storage.Ping(ctx)
}