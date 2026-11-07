package domain

import "time"

// Contact is the complete contact card returned by the API.
type Contact struct {
	ID        int64     `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Telegram  string    `json:"telegram"`
	CompanyID *int64    `json:"company_id"`
	OwnerID   int64     `json:"owner_id"`
	Tags      []string  `json:"tags"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ContactInput contains only editable fields. Update replaces all these fields.
// A pointer distinguishes an omitted owner_id (default 1) from explicit zero.
type ContactInput struct {
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Phone     string   `json:"phone"`
	Email     string   `json:"email"`
	Telegram  string   `json:"telegram"`
	CompanyID *int64   `json:"company_id"`
	OwnerID   *int64   `json:"owner_id"`
	Tags      []string `json:"tags"`
	Comment   string   `json:"comment"`
}
