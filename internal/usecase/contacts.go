package usecase

import (
	"agniscrm/internal/domain"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ContactRepository must validate references and deletion conflicts atomically
// with writes, so concurrent entity operations cannot leave broken references.
type ContactRepository interface {
	ListContacts() []domain.Contact
	GetContact(int64) (domain.Contact, error)
	CreateContact(domain.ContactInput) (domain.Contact, error)
	UpdateContact(int64, domain.ContactInput) (domain.Contact, error)
	DeleteContact(int64) error
}

type Contacts struct{ store ContactRepository }

func NewContacts(store ContactRepository) *Contacts      { return &Contacts{store: store} }
func (s *Contacts) List() []domain.Contact               { return s.store.ListContacts() }
func (s *Contacts) Get(id int64) (domain.Contact, error) { return s.store.GetContact(id) }
func (s *Contacts) Delete(id int64) error                { return s.store.DeleteContact(id) }

func (s *Contacts) Create(in domain.ContactInput) (domain.Contact, error) {
	in, err := normalizeContact(in)
	if err != nil {
		return domain.Contact{}, err
	}
	return s.store.CreateContact(in)
}

func (s *Contacts) Update(id int64, in domain.ContactInput) (domain.Contact, error) {
	in, err := normalizeContact(in)
	if err != nil {
		return domain.Contact{}, err
	}
	return s.store.UpdateContact(id, in)
}

var phonePattern = regexp.MustCompile(`^\+[0-9]{10,15}$`)
var telegramPattern = regexp.MustCompile(`^@[A-Za-z0-9_]{5,32}$`)

func normalizeContact(in domain.ContactInput) (domain.ContactInput, error) {
	in.FirstName = strings.TrimSpace(in.FirstName)
	in.LastName = strings.TrimSpace(in.LastName)
	in.Phone = strings.TrimSpace(in.Phone)
	in.Email = strings.TrimSpace(in.Email)
	in.Telegram = strings.TrimSpace(in.Telegram)
	in.Comment = strings.TrimSpace(in.Comment)
	if in.OwnerID == nil {
		id := int64(1)
		in.OwnerID = &id
	}
	if in.Telegram != "" && !strings.HasPrefix(in.Telegram, "@") {
		in.Telegram = "@" + in.Telegram
	}
	var details []domain.FieldError
	add := func(field, message string) {
		details = append(details, domain.FieldError{Field: field, Message: message})
	}
	length := func(field, value string, min, max int) {
		n := utf8.RuneCountInString(value)
		if n < min {
			add(field, "обязательное поле")
		} else if n > max {
			add(field, "превышена допустимая длина")
		}
	}
	length("first_name", in.FirstName, 1, 100)
	length("last_name", in.LastName, 0, 100)
	length("comment", in.Comment, 0, 2000)
	if in.Phone == "" && in.Email == "" && in.Telegram == "" {
		add("contacts", "укажите телефон, email или Telegram")
	}
	if in.Phone != "" && !phonePattern.MatchString(in.Phone) {
		add("phone", "ожидается + и от 10 до 15 цифр")
	}
	if in.Email != "" {
		address, err := mail.ParseAddress(in.Email)
		if err != nil || address.Address != in.Email || utf8.RuneCountInString(in.Email) > 254 {
			add("email", "некорректный email")
		}
	}
	if in.Telegram != "" && !telegramPattern.MatchString(in.Telegram) {
		add("telegram", "ожидается @ и 5–32 символа: латиница, цифры, подчёркивание")
	}
	if *in.OwnerID <= 0 {
		add("owner_id", "идентификатор должен быть положительным")
	}
	if in.CompanyID != nil && *in.CompanyID <= 0 {
		add("company_id", "идентификатор должен быть положительным")
	}
	// Validate the supplied list before deduplication; limits apply to input too.
	if len(in.Tags) > 10 {
		add("tags", "допускается не более 10 тегов")
	}
	tags := make([]string, 0, len(in.Tags))
	seen := make(map[string]bool)
	for _, tag := range in.Tags {
		tag = strings.TrimSpace(tag)
		if n := utf8.RuneCountInString(tag); n < 1 || n > 30 {
			add("tags", "каждый тег должен содержать 1–30 символов")
		}
		if !seen[tag] {
			tags = append(tags, tag)
			seen[tag] = true
		}
	}
	in.Tags = tags
	if len(details) > 0 {
		return in, &domain.ValidationError{Details: details}
	}
	return in, nil
}
