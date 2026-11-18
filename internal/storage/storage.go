package storage

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"agnis-crm/internal/domain"
	"agnis-crm/internal/storage/memory"
	"agnis-crm/internal/storage/postgres"
)

// Storage — фасад над хранилищем. Делегирует PostgreSQL, когда он подключён,
// иначе работает на моковых данных (этап 1, модульные тесты).
type Storage struct {
	*memory.Store
	pg *postgres.Store
}

// New создаёт хранилище на моковых данных (используется в тестах).
func New() *Storage { return &Storage{Store: memory.NewWithSeed()} }

// NewWithDB создаёт хранилище с подключением к PostgreSQL.
func NewWithDB(db *sqlx.DB, timeout time.Duration) *Storage {
	return &Storage{Store: memory.NewWithSeed(), pg: postgres.New(db, timeout)}
}

// Ping проверяет доступность хранилища. Для мокового хранилища всегда ok.
func (s *Storage) Ping(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	return s.pg.Ping(ctx)
}

// ListContacts делегирует PostgreSQL или моковому хранилищу.
func (s *Storage) ListContacts(ctx context.Context) ([]domain.Contact, error) {
	if s.pg != nil {
		return s.pg.ListContacts(ctx)
	}
	return s.Store.ListContacts(ctx)
}

func (s *Storage) GetContact(ctx context.Context, id int64) (domain.Contact, error) {
	if s.pg != nil {
		return s.pg.GetContact(ctx, id)
	}
	return s.Store.GetContact(ctx, id)
}

func (s *Storage) CreateContact(ctx context.Context, in domain.ContactInput) (domain.Contact, error) {
	if s.pg != nil {
		return s.pg.CreateContact(ctx, in)
	}
	return s.Store.CreateContact(ctx, in)
}

func (s *Storage) UpdateContact(ctx context.Context, id int64, in domain.ContactInput) (domain.Contact, error) {
	if s.pg != nil {
		return s.pg.UpdateContact(ctx, id, in)
	}
	return s.Store.UpdateContact(ctx, id, in)
}

func (s *Storage) DeleteContact(ctx context.Context, id int64) error {
	if s.pg != nil {
		return s.pg.DeleteContact(ctx, id)
	}
	return s.Store.DeleteContact(ctx, id)
}
