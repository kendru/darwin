input "name" {
  type = "string"
  description = "Project name"
  default = "My Project"
}

attribute "slug" {
  val = slugify(var.name)
}

// update_file "main.tf" {
//   source = "../main.tf"
//   skip_if_match = "./modules/${slugify(var.name)}"
//   append = <<EOT-
//   module "${snake_case(var.name)}" {
//     source = "./modules/${slugify(var.name)}"

//     # var1 = "foo"
//     # var2 = 123

//     providers = {
//       aws = aws
//     }
//   }
//   EOT
// }
