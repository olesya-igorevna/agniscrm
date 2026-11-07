package handler_test

import (
	"agniscrm/internal/domain"
	"agniscrm/internal/handler"
	"agniscrm/internal/storage/memory"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

func router() http.Handler {
	gin.SetMode(gin.TestMode)
	return handler.New(memory.NewWithSeed()).Router()
}

func request(t *testing.T, r http.Handler, method, path, body string, want int) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != want {
		t.Fatalf("%s %s: got %d, want %d: %s", method, path, w.Code, want, w.Body.String())
	}
	if !strings.HasPrefix(w.Header().Get("Content-Type"), "application/json") {
		t.Fatalf("non-JSON response: %s", w.Header())
	}
	return w
}

func contact(t *testing.T, w *httptest.ResponseRecorder) domain.Contact {
	t.Helper()
	var c domain.Contact
	if err := json.Unmarshal(w.Body.Bytes(), &c); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestContactLifecycle(t *testing.T) {
	r := router()
	w := request(t, r, "POST", "/api/v1/contacts", `{"first_name":"  Арина  ","last_name":" Федюк ","telegram":" arina_test ","tags":[" новый ","новый","лид"],"company_id":1,"owner_id":3,"id":999,"created_at":"ignored"}`, 201)
	c := contact(t, w)
	if c.ID != 4 || c.FirstName != "Арина" || c.LastName != "Федюк" || c.Telegram != "@arina_test" || len(c.Tags) != 2 || c.CreatedAt.IsZero() || c.CreatedAt.Location().String() != "UTC" {
		t.Fatalf("bad created contact: %+v", c)
	}
	path := fmt.Sprintf("/api/v1/contacts/%d", c.ID)
	got := contact(t, request(t, r, "GET", path, "", 200))
	if got.ID != c.ID || got.Telegram != c.Telegram {
		t.Fatal("created contact was not persisted")
	}
	var list struct {
		Items []domain.Contact `json:"items"`
		Total int              `json:"total"`
	}
	if err := json.Unmarshal(request(t, r, "GET", "/api/v1/contacts", "", 200).Body.Bytes(), &list); err != nil {
		t.Fatal(err)
	}
	if list.Total != 4 || list.Items[3].ID != c.ID {
		t.Fatalf("bad list: %+v", list)
	}
	updated := contact(t, request(t, r, "POST", path+"/update", `{"first_name":"Обновлённое имя","email":"user@example.com"}`, 200))
	if updated.OwnerID != 1 || updated.CompanyID != nil || updated.Telegram != "" || updated.LastName != "" || len(updated.Tags) != 0 || !updated.CreatedAt.Equal(c.CreatedAt) || updated.UpdatedAt.Before(c.UpdatedAt) {
		t.Fatalf("update did not replace editable fields: %+v", updated)
	}
	if !strings.Contains(request(t, r, "GET", path, "", 200).Body.String(), `"tags":[]`) {
		t.Fatal("empty tags must be []")
	}
	w = request(t, r, "POST", path+"/delete", "", 200)
	if !strings.Contains(w.Body.String(), `"deleted":true`) {
		t.Fatal(w.Body.String())
	}
	request(t, r, "GET", path, "", 404)
	request(t, r, "POST", path+"/delete", "", 404)
	next := contact(t, request(t, r, "POST", "/api/v1/contacts", `{"first_name":"Следующий","phone":"+79991234567"}`, 201))
	if next.ID != 5 {
		t.Fatal("deleted IDs must not be reused")
	}
}

func TestContactErrors(t *testing.T) {
	r := router()
	cases := []struct {
		name, method, path, body string
		status                   int
		code, field              string
	}{
		{"missing name", "POST", "/api/v1/contacts", `{"phone":"+79991234567"}`, 400, "validation_error", "first_name"},
		{"blank name", "POST", "/api/v1/contacts", `{"first_name":"   ","email":"a@example.com"}`, 400, "validation_error", "first_name"},
		{"no channel", "POST", "/api/v1/contacts", `{"first_name":"Анна"}`, 400, "validation_error", "contacts"},
		{"phone", "POST", "/api/v1/contacts", `{"first_name":"Анна","phone":"123"}`, 400, "validation_error", "phone"},
		{"email", "POST", "/api/v1/contacts", `{"first_name":"Анна","email":"Name <a@example.com>"}`, 400, "validation_error", "email"},
		{"telegram", "POST", "/api/v1/contacts", `{"first_name":"Анна","telegram":"ab"}`, 400, "validation_error", "telegram"},
		{"tag", "POST", "/api/v1/contacts", `{"first_name":"Анна","telegram":"anna_test","tags":[" "]}`, 400, "validation_error", "tags"},
		{"unknown owner", "POST", "/api/v1/contacts", `{"first_name":"Анна","telegram":"anna_test","owner_id":99}`, 400, "validation_error", "owner_id"},
		{"zero owner", "POST", "/api/v1/contacts", `{"first_name":"Анна","telegram":"anna_test","owner_id":0}`, 400, "validation_error", "owner_id"},
		{"unknown company", "POST", "/api/v1/contacts", `{"first_name":"Анна","telegram":"anna_test","company_id":99}`, 400, "validation_error", "company_id"},
		{"bad JSON", "POST", "/api/v1/contacts", `{"first_name":`, 400, "invalid_json", ""},
		{"wrong type", "POST", "/api/v1/contacts", `{"first_name":123}`, 400, "invalid_json", ""},
		{"null string", "POST", "/api/v1/contacts", `{"first_name":null}`, 400, "invalid_json", ""},
		{"null tag", "POST", "/api/v1/contacts", `{"tags":[null]}`, 400, "invalid_json", ""},
		{"array", "POST", "/api/v1/contacts", `[]`, 400, "invalid_json", ""},
		{"null", "POST", "/api/v1/contacts", `null`, 400, "invalid_json", ""},
		{"two objects", "POST", "/api/v1/contacts", `{} {}`, 400, "invalid_json", ""},
		{"bad id", "GET", "/api/v1/contacts/abc", "", 400, "invalid_id", ""},
		{"zero id", "GET", "/api/v1/contacts/0", "", 400, "invalid_id", ""},
		{"overflow", "GET", "/api/v1/contacts/999999999999999999999", "", 400, "invalid_id", ""},
		{"missing", "GET", "/api/v1/contacts/99", "", 404, "not_found", ""},
		{"update missing", "POST", "/api/v1/contacts/99/update", `{"first_name":"Анна","telegram":"anna_test"}`, 404, "not_found", ""},
		{"full replacement", "POST", "/api/v1/contacts/1/update", `{"comment":"Текст"}`, 400, "validation_error", "first_name"},
		{"in use", "POST", "/api/v1/contacts/1/delete", "", 409, "conflict", ""},
		{"method", "DELETE", "/api/v1/contacts/1", "", 405, "method_not_allowed", ""},
		{"route", "GET", "/unknown", "", 404, "route_not_found", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := request(t, r, tc.method, tc.path, tc.body, tc.status)
			var body struct {
				Error struct {
					Code    string              `json:"code"`
					Details []domain.FieldError `json:"details"`
				} `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error.Code != tc.code {
				t.Fatalf("want %s: %s", tc.code, w.Body.String())
			}
			if tc.field != "" {
				found := false
				for _, d := range body.Error.Details {
					if d.Field == tc.field {
						found = true
					}
				}
				if !found {
					t.Fatalf("missing field %s: %s", tc.field, w.Body.String())
				}
			}
		})
	}
}

func TestUnicodeLimits(t *testing.T) {
	r := router()
	for _, n := range []int{100, 101} {
		body, _ := json.Marshal(map[string]any{"first_name": strings.Repeat("Я", n), "telegram": "anna_test"})
		status := 201
		if n > 100 {
			status = 400
		}
		request(t, r, "POST", "/api/v1/contacts", string(body), status)
	}
}

func TestConcurrentCreatesAndCopies(t *testing.T) {
	s := memory.New(memory.ContactRelations{UserIDs: []int64{1}})
	owner := int64(1)
	in := domain.ContactInput{FirstName: "Анна", Phone: "+79991234567", OwnerID: &owner, Tags: []string{"лид"}}
	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.CreateContact(in); err != nil {
				t.Error(err)
			}
			s.ListContacts()
		}()
	}
	wg.Wait()
	items := s.ListContacts()
	if len(items) != 30 {
		t.Fatalf("lost contacts: %d", len(items))
	}
	for i, c := range items {
		if c.ID != int64(i+1) {
			t.Fatal("IDs are not unique and sorted")
		}
	}
	in.Tags[0] = "изменён"
	items[0].Tags[0] = "подмена"
	stored, err := s.GetContact(1)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Tags[0] != "лид" {
		t.Fatal("storage leaked mutable slice")
	}
}

func TestEmptyList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := handler.New(memory.New(memory.ContactRelations{})).Router()
	w := request(t, r, "GET", "/api/v1/contacts", "", 200)
	if !strings.Contains(w.Body.String(), `"items":[]`) {
		t.Fatal(w.Body.String())
	}
}
