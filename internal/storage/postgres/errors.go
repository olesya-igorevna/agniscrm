package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"agnis-crm/internal/domain"
)

// mapError переводит ошибку PostgreSQL в доменную (таблица 9.1 ТЗ).
func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return domain.ErrTimeout
	}
	var pgErr *pq.Error
	if !errors.As(err, &pgErr) {
		// Не серверная SQL-ошибка (обрыв соединения и т.п.).
		return domain.ErrUnavailable
	}
	switch pgErr.Code {
	case "23505": // нарушение уникальности
		return domain.ErrConflict
	case "23503": // нарушение внешнего ключа: сущность используется
		return domain.ErrConflict
	}
	return err
}

// referenceError превращает код 23503 при создании/изменении в ValidationError
// с полем из имени ограничения, например contacts_company_id_fkey -> company_id.
func referenceError(err error) error {
	if err == nil {
		return nil
	}
	var pgErr *pq.Error
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		return mapError(err)
	}
	field := fieldFromConstraint(pgErr.Constraint)
	return &domain.ValidationError{Details: []domain.FieldError{
		{Field: field, Message: "связанная сущность не найдена"},
	}}
}

// fieldFromConstraint извлекает имя поля из имени ограничения вида
// <table>_<field>_fkey, например contacts_company_id_fkey -> company_id.
// Возвращает "id", если распарсить не удалось.
func fieldFromConstraint(name string) string {
	const suffix = "_fkey"
	if len(name) <= len(suffix) {
		return "id"
	}
	base := name[:len(name)-len(suffix)]
	// Имя таблицы отделяется первым подчёркиванием, поле — всё остальное.
	for i := 0; i < len(base); i++ {
		if base[i] == '_' {
			return base[i+1:]
		}
	}
	return base
}
