terraform {
  required_providers {
    fireworks = {
      source = "registry.terraform.io/amrutp24/fireworks"
    }
  }
}

# Reads FIREWORKS_API_KEY and FIREWORKS_ACCOUNT_ID from the environment.
provider "fireworks" {}

resource "fireworks_deployment" "llama" {
  base_model        = "accounts/fireworks/models/llama-v3p1-8b-instruct"
  display_name      = "example inference"
  min_replica_count = 0 # scale to zero when idle â€” no GPU cost
  max_replica_count = 1
}

resource "fireworks_dataset" "training" {
  dataset_id    = "example-training-data"
  display_name  = "Example training data"
  external_url  = "gs://my-bucket/train.jsonl"
  format        = "CHAT"
  example_count = 1200
}

output "deployment_id" {
  value = fireworks_deployment.llama.id
}

output "deployment_state" {
  value = fireworks_deployment.llama.state
}
