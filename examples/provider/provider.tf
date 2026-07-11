terraform {
  required_providers {
    fireworks = {
      source = "amrutp24/fireworks"
    }
  }
}

provider "fireworks" {
  # api_key and account_id can also come from FIREWORKS_API_KEY
  # and FIREWORKS_ACCOUNT_ID environment variables.
  account_id = "my-account"
}
