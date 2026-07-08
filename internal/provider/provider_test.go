package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/amrutp24/terraform-provider-fireworks/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"fireworks": providerserver.NewProtocol6WithError(New("test")()),
}

// mockFireworks is an in-memory implementation of the Fireworks control-plane
// endpoints the provider uses, so acceptance tests run without credentials.
type mockFireworks struct {
	mu          sync.Mutex
	nextID      int
	deployments map[string]*client.Deployment
	datasets    map[string]*client.Dataset
}

func newMockFireworks(t *testing.T) *httptest.Server {
	m := &mockFireworks{
		deployments: map[string]*client.Deployment{},
		datasets:    map[string]*client.Dataset{},
	}
	srv := httptest.NewServer(m)
	t.Cleanup(srv.Close)
	return srv
}

func (m *mockFireworks) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") != "Bearer test-key" {
		http.Error(w, `{"error":"unauthenticated"}`, http.StatusUnauthorized)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Paths: /v1/accounts/test-account/{deployments|datasets}[/{id}]
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 || parts[0] != "v1" || parts[1] != "accounts" || parts[2] != "test-account" {
		http.NotFound(w, r)
		return
	}
	collection := parts[3]
	id := ""
	if len(parts) > 4 {
		id = parts[4]
	}

	switch collection {
	case "deployments":
		m.serveDeployments(w, r, id)
	case "datasets":
		m.serveDatasets(w, r, id)
	default:
		http.NotFound(w, r)
	}
}

func (m *mockFireworks) serveDeployments(w http.ResponseWriter, r *http.Request, id string) {
	switch {
	case r.Method == http.MethodPost && id == "":
		var dep client.Deployment
		_ = json.NewDecoder(r.Body).Decode(&dep)
		m.nextID++
		depID := fmt.Sprintf("dep%d", m.nextID)
		dep.Name = "accounts/test-account/deployments/" + depID
		dep.State = client.DeploymentStateReady // skip CREATING to keep tests fast
		dep.CreateTime = "2026-07-05T00:00:00Z"
		applyDeploymentDefaults(&dep)
		m.deployments[depID] = &dep
		writeJSON(w, &dep)

	case r.Method == http.MethodGet && id != "":
		dep, ok := m.deployments[id]
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, dep)

	case r.Method == http.MethodPatch && id != "":
		existing, ok := m.deployments[id]
		if !ok {
			http.NotFound(w, r)
			return
		}
		var patch client.Deployment
		_ = json.NewDecoder(r.Body).Decode(&patch)
		patch.Name = existing.Name
		patch.CreateTime = existing.CreateTime
		patch.State = client.DeploymentStateReady
		applyDeploymentDefaults(&patch)
		m.deployments[id] = &patch
		writeJSON(w, &patch)

	case r.Method == http.MethodDelete && id != "":
		if _, ok := m.deployments[id]; !ok {
			http.NotFound(w, r)
			return
		}
		delete(m.deployments, id)
		writeJSON(w, map[string]string{})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// applyDeploymentDefaults mimics documented server-side defaults.
func applyDeploymentDefaults(dep *client.Deployment) {
	if dep.MinReplicaCount == nil {
		zero := int64(0)
		dep.MinReplicaCount = &zero
	}
	if dep.MaxReplicaCount == nil {
		max := *dep.MinReplicaCount
		if max < 1 {
			max = 1
		}
		dep.MaxReplicaCount = &max
	}
	if dep.AcceleratorType == "" {
		dep.AcceleratorType = "NVIDIA_A100_80GB"
	}
	if dep.AcceleratorCount == nil {
		one := int64(1)
		dep.AcceleratorCount = &one
	}
	if dep.Precision == "" {
		dep.Precision = "FP16"
	}
}

func (m *mockFireworks) serveDatasets(w http.ResponseWriter, r *http.Request, id string) {
	switch {
	case r.Method == http.MethodPost && id == "":
		var req struct {
			Dataset   *client.Dataset `json:"dataset"`
			DatasetID string          `json:"datasetId"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		ds := req.Dataset
		if ds == nil {
			ds = &client.Dataset{}
		}
		// Real-API rule (verified 2026-07-07): uploaded datasets (no
		// externalUrl) must declare exampleCount.
		if ds.ExternalURL == "" && ds.ExampleCount == 0 {
			http.Error(w, `{"code":9,"message":"error validating dataset: example_count is required for uploaded datasets"}`, http.StatusBadRequest)
			return
		}
		ds.Name = "accounts/test-account/datasets/" + req.DatasetID
		ds.State = "READY"
		if ds.Format == "" {
			ds.Format = "CHAT"
		}
		m.datasets[req.DatasetID] = ds
		writeJSON(w, ds)

	case r.Method == http.MethodGet && id != "":
		ds, ok := m.datasets[id]
		if !ok {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, ds)

	case r.Method == http.MethodDelete && id != "":
		if _, ok := m.datasets[id]; !ok {
			http.NotFound(w, r)
			return
		}
		delete(m.datasets, id)
		writeJSON(w, map[string]string{})

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// testProviderConfig points the provider at the mock server.
func testProviderConfig(mockURL string) string {
	return fmt.Sprintf(`
provider "fireworks" {
  api_key    = "test-key"
  account_id = "test-account"
  base_url   = %q
}
`, mockURL)
}
