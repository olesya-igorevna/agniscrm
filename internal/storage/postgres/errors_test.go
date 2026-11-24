package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/lib/pq"

	"agnis-crm/internal/domain"
)

// TestMapErrorTable проверяет перевод ошибок БД по таблице 9.1 ТЗ.
func TestMapErrorTable(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want error
	}{
		{"nil", nil, nil},
		{"no rows -> not found", sql.ErrNoRows, domain.ErrNotFound},
		{"deadline -> timeout", context.DeadlineExceeded, domain.ErrTimeout},
		{"unique violation -> conflict", &pq.Error{Code: "23505"}, domain.ErrConflict},
		{"fk violation -> conflict", &pq.Error{Code: "23503"}, domain.ErrConflict},
		{"check violation -> passthrough", &pq.Error{Code: "23514"}, &pq.Error{Code: "23514"}},
		{"non-sql error -> unavailable", errors.New("connection refused"), domain.ErrUnavailable},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := mapError(tc.err)
			if tc.want == nil {
				if got != nil {
					t.Fatalf("mapError(%v) = %v, want nil", tc.err, got)
				}
				return
			}
			// passthrough (неизвестный код) возвращает тот же указатель.
			if pgWant, ok := tc.want.(*pq.Error); ok {
				if got != tc.err {
					t.Fatalf("mapError(%v) = %v, want passthrough %v", tc.err, got, pgWant)
				}
				return
			}
			if !errors.Is(got, tc.want) {
				t.Fatalf("mapError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestReferenceError(t *testing.T) {
	cases := []struct {
		name      string
		err       error
		wantField string
	}{
		{"company fk", &pq.Error{Code: "23503", Constraint: "contacts_company_id_fkey"}, "company_id"},
		{"owner fk", &pq.Error{Code: "23503", Constraint: "contacts_owner_id_fkey"}, "owner_id"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := referenceError(tc.err)
			var validation *domain.ValidationError
			if !errors.As(got, &validation) {
				t.Fatalf("referenceError(%v) = %v, want ValidationError", tc.err, got)
			}
			if len(validation.Details) != 1 || validation.Details[0].Field != tc.wantField {
				t.Fatalf("field = %+v, want %q", validation.Details, tc.wantField)
			}
		})
	}
}

// TestReferenceErrorNotFK проверяет, что не-FK ошибка не превращается в ValidationError.
func TestReferenceErrorNotFK(t *testing.T) {
	got := referenceError(sql.ErrNoRows)
	if !errors.Is(got, domain.ErrNotFound) {
		t.Fatalf("referenceError(sql.ErrNoRows) = %v, want ErrNotFound", got)
	}
}

func TestFieldFromConstraint(t *testing.T) {
	cases := map[string]string{
		"contacts_company_id_fkey": "company_id",
		"contacts_owner_id_fkey":   "owner_id",
		"deals_stage_id_fkey":      "stage_id",
		"tasks_deal_id_fkey":       "deal_id",
		"short":                    "id",
		"_fkey":                    "id",
		"nofield_fkey":             "nofield",
	}
	for in, want := range cases {
		if got := fieldFromConstraint(in); got != want {
			t.Errorf("fieldFromConstraint(%q) = %q, want %q", in, got, want)
		}
	}
}
