package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSecretResource(t *testing.T) {
	srv := newMockFireworks(t)
	mock, ok := srv.Config.Handler.(*mockFireworks)
	if !ok {
		t.Fatalf("unexpected mock handler type %T", srv.Config.Handler)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig(srv.URL) + `
resource "fireworks_secret" "test" {
  secret_id = "wandb"
  key_name  = "WANDB_API_KEY"
  value     = "first-value"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fireworks_secret.test", "id", "wandb"),
					resource.TestCheckResourceAttr("fireworks_secret.test", "name", "accounts/test-account/secrets/wandb"),
					resource.TestCheckResourceAttr("fireworks_secret.test", "key_name", "WANDB_API_KEY"),
					// The API never returns the value; it must survive refresh.
					resource.TestCheckResourceAttr("fireworks_secret.test", "value", "first-value"),
					func(*terraform.State) error { return mockSecretValueIs(mock, "wandb", "first-value") },
				),
			},
			// key_name and value update in place.
			{
				Config: testProviderConfig(srv.URL) + `
resource "fireworks_secret" "test" {
  secret_id = "wandb"
  key_name  = "WANDB_KEY"
  value     = "second-value"
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("fireworks_secret.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fireworks_secret.test", "key_name", "WANDB_KEY"),
					resource.TestCheckResourceAttr("fireworks_secret.test", "value", "second-value"),
					func(*terraform.State) error { return mockSecretValueIs(mock, "wandb", "second-value") },
				),
			},
			// Changing secret_id replaces.
			{
				Config: testProviderConfig(srv.URL) + `
resource "fireworks_secret" "test" {
  secret_id = "wandb-v2"
  key_name  = "WANDB_KEY"
  value     = "second-value"
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("fireworks_secret.test", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.TestCheckResourceAttr("fireworks_secret.test", "id", "wandb-v2"),
			},
			{
				ResourceName:            "fireworks_secret.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"value"},
			},
		},
	})
}

func mockSecretValueIs(m *mockFireworks, id, want string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.secrets[id]
	if !ok {
		return fmt.Errorf("secret %s not found in mock", id)
	}
	if s.Value != want {
		return fmt.Errorf("mock secret %s has value %q, want %q", id, s.Value, want)
	}
	return nil
}
