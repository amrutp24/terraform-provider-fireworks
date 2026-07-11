resource "fireworks_dataset" "training" {
  dataset_id    = "training-data-v1"
  display_name  = "Training data v1"
  external_url  = "gs://my-bucket/train.jsonl"
  format        = "CHAT"
  example_count = 1200
}
