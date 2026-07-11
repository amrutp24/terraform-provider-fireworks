resource "fireworks_deployment" "llama" {
  base_model        = "accounts/fireworks/models/llama-v3p1-8b-instruct"
  display_name      = "prod inference"
  min_replica_count = 0 # scale to zero when idle
  max_replica_count = 2
  accelerator_type  = "NVIDIA_H100_80GB"
  precision         = "FP8"
  region            = "US_IOWA_1"
}
