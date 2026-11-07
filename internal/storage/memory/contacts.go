package memory

import (
	"agniscrm/internal/domain"
	"sort"
	"sync"
	"time"
)

// ContactRelations is a temporary fixture for entities owned by other teammates.
// Replace these ID sets with the shared entity maps when integrating their code.
type ContactRelations struct {
	UserIDs        []int64
	CompanyIDs     []int64
	DealContactIDs []int64
	TaskContactIDs []int64
}

type Store struct {
	mu                             sync.RWMutex
	contacts                       map[int64]domain.Contact
	nextContactID                  int64
	users, companies, usedContacts map[int64]bool
}

func New(rel ContactRelations) *Store {
	s := &Store{contacts: make(map[int64]domain.Contact), nextContactID: 1,
		users: make(map[int64]bool), companies: make(map[int64]bool), usedContacts: make(map[int64]bool)}
	for _, id := range rel.UserIDs {
		s.users[id] = true
	}
	for _, id := range rel.CompanyIDs {
		s.companies[id] = true
	}
	for _, id := range rel.DealContactIDs {
		s.usedContacts[id] = true
	}
	for _, id := range rel.TaskContactIDs {
		s.usedContacts[id] = true
	}
	return s
}

func NewWithSeed() *Store {
	s := New(ContactRelations{UserIDs: []int64{1, 2, 3}, CompanyIDs: []int64{1, 2}, DealContactIDs: []int64{1, 2, 3}})
	company1, company2, owner3, owner2 := int64(1), int64(2), int64(3), int64(2)
	seed := []domain.ContactInput{
		{FirstName: "Иван", LastName: "Петров", Phone: "+79991234567", CompanyID: &company1, OwnerID: &owner3},
		{FirstName: "Ольга", LastName: "Соколова", Phone: "+79161112233", CompanyID: &company2, OwnerID: &owner3},
		// Appendix A omits a contact channel for Sergey; do not invent one.
		{FirstName: "Сергей", LastName: "Лебедев", OwnerID: &owner2},
	}
	for _, in := range seed {
		in.Tags = []string{}
		now := time.Now().UTC()
		c := contactFromInput(in)
		c.ID, c.CreatedAt, c.UpdatedAt = s.nextContactID, now, now
		s.contacts[c.ID] = c
		s.nextContactID++
	}
	return s
}

// All returned records are independent copies, including slices and pointers.
func clone(c domain.Contact) domain.Contact {
	c.Tags = append([]string{}, c.Tags...)
	if c.CompanyID != nil {
		id := *c.CompanyID
		c.CompanyID = &id
	}
	return c
}

func contactFromInput(in domain.ContactInput) domain.Contact {
	c := domain.Contact{FirstName: in.FirstName, LastName: in.LastName, Phone: in.Phone,
		Email: in.Email, Telegram: in.Telegram, CompanyID: in.CompanyID, OwnerID: *in.OwnerID, Tags: in.Tags, Comment: in.Comment}
	return clone(c)
}

func (s *Store) ListContacts() []domain.Contact {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]domain.Contact, 0, len(s.contacts))
	for _, c := range s.contacts {
		items = append(items, clone(c))
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items
}

func (s *Store) GetContact(id int64) (domain.Contact, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.contacts[id]
	if !ok {
		return domain.Contact{}, domain.ErrNotFound
	}
	return clone(c), nil
}

// Caller holds mu during reference checks and mutation.
func (s *Store) checkReferences(in domain.ContactInput) error {
	var details []domain.FieldError
	if in.OwnerID == nil || !s.users[*in.OwnerID] {
		details = append(details, domain.FieldError{Field: "owner_id", Message: "пользователь не найден"})
	}
	if in.CompanyID != nil && !s.companies[*in.CompanyID] {
		details = append(details, domain.FieldError{Field: "company_id", Message: "компания не найдена"})
	}
	if len(details) > 0 {
		return &domain.ValidationError{Details: details}
	}
	return nil
}

func (s *Store) CreateContact(in domain.ContactInput) (domain.Contact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.checkReferences(in); err != nil {
		return domain.Contact{}, err
	}
	now := time.Now().UTC()
	c := contactFromInput(in)
	c.ID, c.CreatedAt, c.UpdatedAt = s.nextContactID, now, now
	s.nextContactID++
	s.contacts[c.ID] = c
	return clone(c), nil
}

func (s *Store) UpdateContact(id int64, in domain.ContactInput) (domain.Contact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.contacts[id]
	if !ok {
		return domain.Contact{}, domain.ErrNotFound
	}
	if err := s.checkReferences(in); err != nil {
		return domain.Contact{}, err
	}
	c := contactFromInput(in)
	c.ID, c.CreatedAt, c.UpdatedAt = id, old.CreatedAt, time.Now().UTC()
	s.contacts[id] = c
	return clone(c), nil
}

func (s *Store) DeleteContact(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.contacts[id]; !ok {
		return domain.ErrNotFound
	}
	if s.usedContacts[id] {
		return domain.ErrConflict
	}
	delete(s.contacts, id)
	return nil
}
