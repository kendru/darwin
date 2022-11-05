input "name" {
  type = "string"
  description = "Module name"
  default = "My Module"
}

input "region" {
  type = "string"
  description = "AWS Region"
  default = "us-east-1"
}

attribute "slug" {
  val = slugify(var.name)
}

attribute "key" {
  val = snake_case(var.name)
}

attribute "email" {
  val = gitconfig("user.email")
}
