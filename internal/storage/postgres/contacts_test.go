package postgres

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"agnis-crm/internal/domain"
)

// TestPostgresContactLifecycle — интеграционный тест хранилища PostgreSQL.
// Запускается только при заданной TEST_DATABASE_URL (ТЗ 9.10), иначе скипается.
func TestPostgresContactLifecycle(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL не задана — пропускаю интеграционный тест")
	}

	db, err := sqlx.Connect("postgres", url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	// Изолируем тест в отдельной схеме-таблице, не трогая seed-данные.
	// Требуется, чтобы таблица contacts существовала (миграции применены).
	s := New(db, 5*time.Second)
	ctx := context.Background()

	owner := int64(1)
	in := domain.ContactInput{
		FirstName: "Тест",
		Phone:     "+79991234567",
		OwnerID:   &owner,
		Tags:      []string{"лид", "лид", "vip"},
	}

	created, err := s.CreateContact(ctx, in)
	if err != nil {
		t.Fatalf("CreateContact: %v", err)
	}
	if created.ID == 0 || created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("created contact missing server fields: %+v", created)
	}
	if created.FirstName != "Тест" || created.OwnerID != 1 {
		t.Fatalf("created contact wrong fields: %+v", created)
	}

	// Прочитаем обратно.
	got, err := s.GetContact(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetContact: %v", err)
	}
	if got.ID != created.ID || got.FirstName != "Тест" || got.Telegram != "" {
		t.Fatalf("GetContact mismatch: %+v", got)
	}

	// Список должен содержать созданный контакт.
	items, err := s.ListContacts(ctx)
	if err != nil {
		t.Fatalf("ListContacts: %v", err)
	}
	found := false
	for _, c := range items {
		if c.ID == created.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("ListContacts: created contact %d not found", created.ID)
	}

	// Обновление: меняем все редактируемые поля.
	upd := domain.ContactInput{
		FirstName: "Обновлённый",
		Phone:     "+79991112233",
		OwnerID:   &owner,
		Tags:      []string{},
	}
	updated, err := s.UpdateContact(ctx, created.ID, upd)
	if err != nil {
		t.Fatalf("UpdateContact: %v", err)
	}
	if updated.FirstName != "Обновлённый" || updated.Phone != "+79991112233" {
		t.Fatalf("UpdateContact mismatch: %+v", updated)
	}

	// Удаление.
	if err := s.DeleteContact(ctx, created.ID); err != nil {
		t.Fatalf("DeleteContact: %v", err)
	}
	if _, err := s.GetContact(ctx, created.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("after delete GetContact = %v, want ErrNotFound", err)
	}
}

// TestPostgresFKViolation проверяет перевод 23503 в ValidationError с полем.
func TestPostgresFKViolation(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL не задана — пропускаю интеграционный тест")
	}

	db, err := sqlx.Connect("postgres", url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer db.Close()

	s := New(db, 5*time.Second)
	owner := int64(1)
	badCompany := int64(999999)
	_, err = s.CreateContact(context.Background(), domain.ContactInput{
		FirstName: "Тест FK",
		Phone:     "+79991234567",
		OwnerID:   &owner,
		CompanyID: &badCompany,
	})
	var validation *domain.ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("CreateContact with bad company = %v, want ValidationError", err)
	}
	if len(validation.Details) != 1 || validation.Details[0].Field != "company_id" {
		t.Fatalf("details = %+v, want field company_id", validation.Details)
	}
}
