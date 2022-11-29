terraform {
  required_providers {
    github = {
      source  = "integrations/github"
      version = "~> 4.0"
    }
  }
}

variable "repo_name" {
 type        = string
}

variable "github_token" {
 type        = string
}

variable "github_user" {
 type        = string
}

# Configure the GitHub Provider

provider "github" {
  token = var.github_token
  owner = var.github_user
}



resource "github_repository" "github-repo" {
  name        = var.repo_name
  visibility = "public"
  auto_init = true
}
