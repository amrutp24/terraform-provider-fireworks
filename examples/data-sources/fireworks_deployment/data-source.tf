data "fireworks_deployment" "existing" {
  id = "abc123"
}

output "deployment_state" {
  value = data.fireworks_deployment.existing.state
}
