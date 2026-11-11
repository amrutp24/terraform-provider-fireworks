package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

const testDeployedModelBase = `
resource "fireworks_deployment" "base" {
  base_model = "accounts/fireworks/models/llama-v3p1-8b-instruct"
}
`

func TestAccDeployedModelResource(t *testing.T) {
	mock := newMockFireworks(t)
	const addr = "fireworks_deployed_model.lora"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create on a dedicated deployment with defaults.
			{
				Config: testProviderConfig(mock.URL) + testDeployedModelBase + `
resource "fireworks_deployed_model" "lora" {
  model      = "accounts/test-account/models/lora-a"
  deployment = fireworks_deployment.base.name
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(addr, "id"),
					resource.TestCheckResourceAttrPair(addr, "deployment", "fireworks_deployment.base", "name"),
					resource.TestCheckResourceAttr(addr, "state", "DEPLOYED"),
					resource.TestCheckResourceAttr(addr, "default", "true"),
					resource.TestCheckResourceAttr(addr, "serverless", "false"),
					resource.TestCheckResourceAttr(addr, "public", "false"),
					resource.TestCheckNoResourceAttr(addr, "display_name"),
				),
			},
			// In-place update of mutable fields.
			{
				Config: testProviderConfig(mock.URL) + testDeployedModelBase + `
resource "fireworks_deployed_model" "lora" {
  model        = "accounts/test-account/models/lora-a"
  deployment   = fireworks_deployment.base.name
  display_name = "lora a"
  description  = "first adapter"
  public       = true
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "display_name", "lora a"),
					resource.TestCheckResourceAttr(addr, "description", "first adapter"),
					resource.TestCheckResourceAttr(addr, "public", "true"),
					resource.TestCheckResourceAttr(addr, "default", "true"),
				),
			},
			// Changing model forces replacement.
			{
				Config: testProviderConfig(mock.URL) + testDeployedModelBase + `
resource "fireworks_deployed_model" "lora" {
  model        = "accounts/test-account/models/lora-b"
  deployment   = fireworks_deployment.base.name
  display_name = "lora a"
  description  = "first adapter"
  public       = true
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(addr, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "model", "accounts/test-account/models/lora-b"),
					resource.TestCheckResourceAttr(addr, "state", "DEPLOYED"),
				),
			},
			// Import.
			{
				ResourceName:            addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_ready", "replace_merged_addon"},
			},
		},
	})
}

func TestAccDeployedModelResourceServerless(t *testing.T) {
	mock := newMockFireworks(t)
	const addr = "fireworks_deployed_model.lora"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig(mock.URL) + `
resource "fireworks_deployed_model" "lora" {
  model      = "accounts/test-account/models/lora-a"
  serverless = true
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "serverless", "true"),
					resource.TestCheckResourceAttr(addr, "default", "true"),
					resource.TestCheckResourceAttr(addr, "state", "DEPLOYED"),
					resource.TestCheckNoResourceAttr(addr, "deployment"),
				),
			},
		},
	})
}

func TestAccDeployedModelResourceTargetValidation(t *testing.T) {
	mock := newMockFireworks(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig(mock.URL) + `
resource "fireworks_deployed_model" "lora" {
  model      = "accounts/test-account/models/lora-a"
  deployment = "accounts/test-account/deployments/dep1"
  serverless = true
}`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`not both`),
			},
			{
				Config: testProviderConfig(mock.URL) + `
resource "fireworks_deployed_model" "lora" {
  model = "accounts/test-account/models/lora-a"
}`,
				PlanOnly:    true,
				ExpectError: regexp.MustCompile(`Set either deployment or serverless`),
			},
		},
	})
}
