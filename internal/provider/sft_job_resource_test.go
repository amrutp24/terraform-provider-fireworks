package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSFTJobResource(t *testing.T) {
	mock := newMockFireworks(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testProviderConfig(mock.URL) + `
resource "fireworks_supervised_fine_tuning_job" "test" {
  display_name  = "tune llama"
  dataset       = "accounts/test-account/datasets/training-data"
  base_model    = "accounts/fireworks/models/llama-v3p1-8b-instruct"
  epochs        = 2
  learning_rate = 0.0001
  lora_rank     = 8
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("fireworks_supervised_fine_tuning_job.test", "id"),
					resource.TestCheckResourceAttr("fireworks_supervised_fine_tuning_job.test", "state", "JOB_STATE_RUNNING"),
					resource.TestCheckResourceAttr("fireworks_supervised_fine_tuning_job.test", "epochs", "2"),
					resource.TestCheckResourceAttr("fireworks_supervised_fine_tuning_job.test", "learning_rate", "0.0001"),
					resource.TestCheckResourceAttr("fireworks_supervised_fine_tuning_job.test", "lora_rank", "8"),
					resource.TestCheckResourceAttrSet("fireworks_supervised_fine_tuning_job.test", "output_model"),
				),
			},
			// Import.
			{
				ResourceName:      "fireworks_supervised_fine_tuning_job.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSFTJobResource_baseModelXorWarmStart(t *testing.T) {
	mock := newMockFireworks(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Neither base_model nor warm_start_from → error.
			{
				Config: testProviderConfig(mock.URL) + `
resource "fireworks_supervised_fine_tuning_job" "test" {
  dataset = "accounts/test-account/datasets/training-data"
}`,
				ExpectError: regexp.MustCompile(`Exactly one of base_model or warm_start_from`),
			},
			// Both set → error.
			{
				Config: testProviderConfig(mock.URL) + `
resource "fireworks_supervised_fine_tuning_job" "test" {
  dataset         = "accounts/test-account/datasets/training-data"
  base_model      = "accounts/fireworks/models/llama-v3p1-8b-instruct"
  warm_start_from = "accounts/test-account/models/previous-tune"
}`,
				ExpectError: regexp.MustCompile(`Exactly one of base_model or warm_start_from`),
			},
		},
	})
}

func TestAccModelDataSource(t *testing.T) {
	mock := newMockFireworks(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig(mock.URL) + `
data "fireworks_model" "test" {
  id = "llama-test"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.fireworks_model.test", "name", "accounts/test-account/models/llama-test"),
					resource.TestCheckResourceAttr("data.fireworks_model.test", "kind", "HF_BASE_MODEL"),
					resource.TestCheckResourceAttr("data.fireworks_model.test", "state", "READY"),
				),
			},
		},
	})
}
