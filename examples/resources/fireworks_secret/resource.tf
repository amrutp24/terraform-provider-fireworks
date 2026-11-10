variable "wandb_api_key" {
  type      = string
  sensitive = true
}

resource "fireworks_secret" "wandb" {
  secret_id = "wandb"
  key_name  = "WANDB_API_KEY"
  value     = var.wandb_api_key
}
