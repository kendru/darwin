provider "aws" {
  region  = "{{.region}}"
  profile = local.aws_profile

  default_tags {
    tags = local.default_tags
  }
}
