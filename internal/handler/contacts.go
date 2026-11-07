package handler

import (
	"agniscrm/internal/domain"
	"bytes"
	"encoding/json"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
)

func contactID(c *gin.Context) (int64, bool) {
	raw := c.Param("id")
	// Reject signs, decimals and overflow, rather than accepting a partial ID.
	for _, ch := range raw {
		if ch < '0' || ch > '9' {
			writeError(c, 400, "invalid_id", "id должен быть положительным целым числом")
			return 0, false
		}
	}
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		writeError(c, 400, "invalid_id", "id должен быть положительным целым числом")
		return 0, false
	}
	return id, true
}

func contactInput(c *gin.Context) (domain.ContactInput, bool) {
	var in domain.ContactInput
	invalid := func() (domain.ContactInput, bool) {
		writeError(c, 400, "invalid_json", "ожидается один JSON-объект с корректными типами полей")
		return in, false
	}
	decoder := json.NewDecoder(c.Request.Body)
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return invalid()
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return invalid()
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return invalid()
	}
	// encoding/json accepts null for a string without an error. The contract
	// uses empty strings instead; only company_id is a nullable editable field.
	for _, key := range []string{"first_name", "last_name", "phone", "email", "telegram", "owner_id", "tags", "comment"} {
		if value, ok := fields[key]; ok && bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return invalid()
		}
	}
	if value, ok := fields["tags"]; ok {
		var tags []json.RawMessage
		if err := json.Unmarshal(value, &tags); err != nil {
			return invalid()
		}
		for _, tag := range tags {
			if bytes.Equal(bytes.TrimSpace(tag), []byte("null")) {
				return invalid()
			}
		}
	}
	// Unknown and server-owned fields are ignored, as required by the TZ.
	if err := json.Unmarshal(raw, &in); err != nil {
		return invalid()
	}
	return in, true
}

func (h *Handler) ListContacts(c *gin.Context) {
	items := h.contacts.List()
	c.JSON(200, gin.H{"items": items, "total": len(items)})
}

func (h *Handler) GetContact(c *gin.Context) {
	id, ok := contactID(c)
	if !ok {
		return
	}
	contact, err := h.contacts.Get(id)
	if err != nil {
		domainError(c, err)
		return
	}
	c.JSON(200, contact)
}

func (h *Handler) CreateContact(c *gin.Context) {
	in, ok := contactInput(c)
	if !ok {
		return
	}
	contact, err := h.contacts.Create(in)
	if err != nil {
		domainError(c, err)
		return
	}
	c.JSON(201, contact)
}

func (h *Handler) UpdateContact(c *gin.Context) {
	id, ok := contactID(c)
	if !ok {
		return
	}
	in, ok := contactInput(c)
	if !ok {
		return
	}
	contact, err := h.contacts.Update(id, in)
	if err != nil {
		domainError(c, err)
		return
	}
	c.JSON(200, contact)
}

func (h *Handler) DeleteContact(c *gin.Context) {
	id, ok := contactID(c)
	if !ok {
		return
	}
	if err := h.contacts.Delete(id); err != nil {
		domainError(c, err)
		return
	}
	c.JSON(200, gin.H{"id": id, "deleted": true})
}
