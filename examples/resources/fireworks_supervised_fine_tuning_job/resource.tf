resource "fireworks_supervised_fine_tuning_job" "tune" {
  display_name = "llama tune v1"
  base_model   = "accounts/fireworks/models/llama-v3p1-8b-instruct"
  dataset      = fireworks_dataset.training.name

  epochs        = 2
  learning_rate = 0.0001
  lora_rank     = 16
}
