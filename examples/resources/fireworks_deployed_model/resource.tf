# Serve the output of a fine-tuning job on serverless.
resource "fireworks_deployed_model" "tuned" {
  model        = fireworks_supervised_fine_tuning_job.tune.output_model
  serverless   = true
  display_name = "llama tune v1"
}

# Or load it onto a dedicated deployment that has addons enabled.
resource "fireworks_deployed_model" "dedicated" {
  model      = "accounts/my-account/models/my-lora"
  deployment = "accounts/my-account/deployments/my-addons-deployment"
  default    = true
}
