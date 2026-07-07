package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDeploymentResource(t *testing.T) {
	mock := newMockFireworks(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with defaults.
			{
				Config: testProviderConfig(mock.URL) + `
resource "fireworks_deployment" "test" {
  base_model = "accounts/fireworks/models/llama-v3p1-8b-instruct"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fireworks_deployment.test", "id"),
					resource.TestCheckResourceAttr("fireworks_deployment.test", "state", "READY"),
					resource.TestCheckResourceAttr("fireworks_deployment.test", "min_replica_count", "0"),
					resource.TestCheckResourceAttr("fireworks_deployment.test", "max_replica_count", "1"),
					resource.TestCheckResourceAttr("fireworks_deployment.test", "accelerator_type", "NVIDIA_A100_80GB"),
					resource.TestCheckResourceAttr("fireworks_deployment.test", "precision", "FP16"),
				),
			},
			// In-place update of scaling and hardware settings.
			{
				Config: testProviderConfig(mock.URL) + `
resource "fireworks_deployment" "test" {
  base_model        = "accounts/fireworks/models/llama-v3p1-8b-instruct"
  display_name      = "prod inference"
  min_replica_count = 1
  max_replica_count = 3
  accelerator_type  = "NVIDIA_H100_80GB"
  precision         = "FP8"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fireworks_deployment.test", "display_name", "prod inference"),
					resource.TestCheckResourceAttr("fireworks_deployment.test", "min_replica_count", "1"),
					resource.TestCheckResourceAttr("fireworks_deployment.test", "max_replica_count", "3"),
					resource.TestCheckResourceAttr("fireworks_deployment.test", "accelerator_type", "NVIDIA_H100_80GB"),
					resource.TestCheckResourceAttr("fireworks_deployment.test", "precision", "FP8"),
				),
			},
			// Import.
			{
				ResourceName:            "fireworks_deployment.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_ready"},
			},
		},
	})
}

func TestAccDeploymentDataSource(t *testing.T) {
	mock := newMockFireworks(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig(mock.URL) + `
resource "fireworks_deployment" "test" {
  base_model = "accounts/fireworks/models/llama-v3p1-8b-instruct"
}

data "fireworks_deployment" "lookup" {
  id = fireworks_deployment.test.id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fireworks_deployment.lookup", "base_model", "accounts/fireworks/models/llama-v3p1-8b-instruct"),
					resource.TestCheckResourceAttr("data.fireworks_deployment.lookup", "state", "READY"),
				),
			},
		},
	})
}
