# Look up a fine-tune's output model, then serve it.
data "fireworks_model" "tuned" {
  id = "my-fine-tuned-model"
}

resource "fireworks_deployment" "tuned" {
  base_model        = data.fireworks_model.tuned.name
  min_replica_count = 0
  max_replica_count = 1
}
