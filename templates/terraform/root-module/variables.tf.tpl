# Global variables.

# Environment-specific variables.

variable "aws_profiles" {
  type        = map(string)
  description = "Name of AWS profile to use for execution. This assumes that the shared credentials mechanism is used."
  default     = {
    dev = "dev-profile-name"
    prod = "prod-profile-name"
  }
}

locals {
  environment = terraform.workspace
  aws_profile = lookup(var.aws_profiles, terraform.workspace)
  default_tags = {
    "application": "{{.slug}}"
    "environment": terraform.workspace
  }
}
