package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatasetResource(t *testing.T) {
	mock := newMockFireworks(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testProviderConfig(mock.URL) + `
resource "fireworks_dataset" "test" {
  dataset_id    = "training-data"
  display_name  = "Training data"
  external_url  = "gs://my-bucket/train.jsonl"
  format        = "CHAT"
  example_count = 1200
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fireworks_dataset.test", "id", "training-data"),
					resource.TestCheckResourceAttr("fireworks_dataset.test", "name", "accounts/test-account/datasets/training-data"),
					resource.TestCheckResourceAttr("fireworks_dataset.test", "state", "READY"),
					resource.TestCheckResourceAttr("fireworks_dataset.test", "example_count", "1200"),
				),
			},
			// Changing dataset_id must force replacement.
			{
				Config: testProviderConfig(mock.URL) + `
resource "fireworks_dataset" "test" {
  dataset_id    = "training-data-v2"
  display_name  = "Training data"
  external_url  = "gs://my-bucket/train.jsonl"
  format        = "CHAT"
  example_count = 1200
}`,
				Check: resource.TestCheckResourceAttr("fireworks_dataset.test", "id", "training-data-v2"),
			},
		},
	})
}
