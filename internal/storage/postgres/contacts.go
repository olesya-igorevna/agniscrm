package postgres

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"agnis-crm/internal/domain"
)

// Store — хранилище PostgreSQL. Реализует те же интерфейсы, что и memory,
// поэтому handler и usecase при переходе не меняются.
type Store struct {
	db      *sqlx.DB
	timeout time.Duration
}

// New создаёт хранилище PostgreSQL с таймаутом на каждый запрос.
func New(db *sqlx.DB, timeout time.Duration) *Store {
	return &Store{db: db, timeout: timeout}
}

// Ping проверяет доступность базы данных.
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// contactRow — внутренняя структура строки таблицы (теги db).
type contactRow struct {
	ID        int64          `db:"id"`
	FirstName string         `db:"first_name"`
	LastName  string         `db:"last_name"`
	Phone     string         `db:"phone"`
	Email     string         `db:"email"`
	Telegram  string         `db:"telegram"`
	CompanyID *int64         `db:"company_id"`
	OwnerID   int64          `db:"owner_id"`
	Tags      pq.StringArray `db:"tags"`
	Comment   string         `db:"comment"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

func (r contactRow) toDomain() domain.Contact {
	return domain.Contact{
		ID:        r.ID,
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Phone:     r.Phone,
		Email:     r.Email,
		Telegram:  r.Telegram,
		CompanyID: r.CompanyID,
		OwnerID:   r.OwnerID,
		Tags:      []string(r.Tags),
		Comment:   r.Comment,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

const contactColumns = `id, first_name, last_name, phone, email, telegram,
	company_id, owner_id, tags, comment, created_at, updated_at`

// withTimeout оборачивает контекст таймаутом и возвращает cancel.
func (s *Store) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, s.timeout)
}

// tagsParam оборачивает слайс тегов для записи. pq.Array(nil) дал бы SQL NULL,
// а колонка tags — NOT NULL; пустой слайс записывается как пустой массив.
func tagsParam(tags []string) any {
	if tags == nil {
		return pq.Array([]string{})
	}
	return pq.Array(tags)
}

func (s *Store) ListContacts(ctx context.Context) ([]domain.Contact, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	var rows []contactRow
	if err := s.db.SelectContext(ctx, &rows, `SELECT `+contactColumns+` FROM contacts ORDER BY id`); err != nil {
		return nil, mapError(err)
	}
	items := make([]domain.Contact, 0, len(rows))
	for _, r := range rows {
		items = append(items, r.toDomain())
	}
	return items, nil
}

func (s *Store) GetContact(ctx context.Context, id int64) (domain.Contact, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	var row contactRow
	err := s.db.GetContext(ctx, &row, `SELECT `+contactColumns+` FROM contacts WHERE id = $1`, id)
	if err != nil {
		return domain.Contact{}, mapError(err)
	}
	return row.toDomain(), nil
}

func (s *Store) CreateContact(ctx context.Context, in domain.ContactInput) (domain.Contact, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	var row contactRow
	err := s.db.GetContext(ctx, &row, `
		INSERT INTO contacts (first_name, last_name, phone, email, telegram,
		                      company_id, owner_id, tags, comment)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING `+contactColumns,
		in.FirstName, in.LastName, in.Phone, in.Email, in.Telegram,
		in.CompanyID, in.OwnerID, tagsParam(in.Tags), in.Comment)
	if err != nil {
		return domain.Contact{}, referenceError(err)
	}
	return row.toDomain(), nil
}

func (s *Store) UpdateContact(ctx context.Context, id int64, in domain.ContactInput) (domain.Contact, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	var row contactRow
	err := s.db.GetContext(ctx, &row, `
		UPDATE contacts
		SET first_name = $1, last_name = $2, phone = $3, email = $4, telegram = $5,
		    company_id = $6, owner_id = $7, tags = $8, comment = $9,
		    updated_at = now()
		WHERE id = $10
		RETURNING `+contactColumns,
		in.FirstName, in.LastName, in.Phone, in.Email, in.Telegram,
		in.CompanyID, in.OwnerID, tagsParam(in.Tags), in.Comment, id)
	if err != nil {
		return domain.Contact{}, referenceError(err)
	}
	return row.toDomain(), nil
}

func (s *Store) DeleteContact(ctx context.Context, id int64) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	res, err := s.db.ExecContext(ctx, `DELETE FROM contacts WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return mapError(err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
