package provider

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/amrutp24/terraform-provider-fireworks/internal/client"
)

// serveSecrets mirrors the real API: the client picks the ID through the
// name field, and value is accepted on create and update but never returned.
func (m *mockFireworks) serveSecrets(w http.ResponseWriter, r *http.Request, id string) {
	if m.secrets == nil {
		m.secrets = map[string]*client.Secret{}
	}
	public := func(s *client.Secret) *client.Secret {
		return &client.Secret{Name: s.Name, KeyName: s.KeyName}
	}

	switch {
	case r.Method == http.MethodPost && id == "":
		var s client.Secret
		_ = json.NewDecoder(r.Body).Decode(&s)
		if !strings.HasPrefix(s.Name, "accounts/test-account/secrets/") || s.KeyName == "" {
			http.Error(w, `{"code":3,"message":"name and keyName are required"}`, http.StatusBadRequest)
			return
		}
		secretID := client.ResourceID(s.Name)
		if _, exists := m.secrets[secretID]; exists {
			http.Error(w, `{"code":6,"message":"secret already exists"}`, http.StatusConflict)
			return
		}
		m.secrets[secretID] = &s
		writeJSON(w, public(&s))

	case r.Method == http.MethodGet && id != "":
		s, ok := m.secrets[id]
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, public(s))

	case r.Method == http.MethodPatch && id != "":
		s, ok := m.secrets[id]
		if !ok {
			http.NotFound(w, r)
			return
		}
		var patch client.Secret
		_ = json.NewDecoder(r.Body).Decode(&patch)
		if patch.KeyName == "" {
			http.Error(w, `{"code":3,"message":"keyName is required"}`, http.StatusBadRequest)
			return
		}
		s.KeyName = patch.KeyName
		if patch.Value != "" {
			s.Value = patch.Value
		}
		writeJSON(w, public(s))

	case r.Method == http.MethodDelete && id != "":
		if _, ok := m.secrets[id]; !ok {
			http.NotFound(w, r)
			return
		}
		delete(m.secrets, id)
		writeJSON(w, map[string]string{})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
