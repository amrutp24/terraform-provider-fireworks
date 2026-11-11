package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/amrutp24/terraform-provider-fireworks/internal/client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

// Destroying a running job must cancel it first: the mock rejects DELETE on a
// job that is still active.
func TestAccSFTJobResource_destroyCancelsRunningJob(t *testing.T) {
	mock := newMockFireworks(t)
	config := testProviderConfig(mock.URL) + `
resource "fireworks_supervised_fine_tuning_job" "test" {
  dataset    = "accounts/test-account/datasets/training-data"
  base_model = "accounts/fireworks/models/llama-v3p1-8b-instruct"
}`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy: func(s *terraform.State) error {
			rs, ok := s.RootModule().Resources["fireworks_supervised_fine_tuning_job.test"]
			if !ok {
				return fmt.Errorf("job not found in pre-destroy state")
			}
			c := client.New(mock.URL, "test-account", "test-key")
			_, err := c.GetSupervisedFineTuningJob(context.Background(), rs.Primary.ID)
			if !client.IsNotFound(err) {
				return fmt.Errorf("expected job %s to be deleted, got err=%v", rs.Primary.ID, err)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttr("fireworks_supervised_fine_tuning_job.test", "state", "JOB_STATE_RUNNING"),
			},
			{
				Config:  config,
				Destroy: true,
			},
		},
	})
}

// A failed cancel on a job that is still running must fail the destroy and
// keep the job in state, even when the cancel call itself returns 404.
func TestAccSFTJobResource_destroyFailsWhenCancelFails(t *testing.T) {
	mock := newMockFireworks(t)
	var breakCancel atomic.Bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if breakCancel.Load() && strings.HasSuffix(r.URL.Path, ":cancel") {
			http.NotFound(w, r)
			return
		}
		mock.Config.Handler.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)

	config := testProviderConfig(srv.URL) + `
resource "fireworks_supervised_fine_tuning_job" "test" {
  dataset    = "accounts/test-account/datasets/training-data"
  base_model = "accounts/fireworks/models/llama-v3p1-8b-instruct"
}`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
			{
				PreConfig:   func() { breakCancel.Store(true) },
				Config:      config,
				Destroy:     true,
				ExpectError: regexp.MustCompile(`may still be running`),
			},
			// The job is still in state and still running; restore cancel so
			// the final destroy can clean up.
			{
				PreConfig: func() { breakCancel.Store(false) },
				Config:    config,
				Check:     resource.TestCheckResourceAttr("fireworks_supervised_fine_tuning_job.test", "state", "JOB_STATE_RUNNING"),
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
