terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "~> 6.0"
    }

    http = {
      source = "hashicorp/http"
      version = "3.5.0"
    }

    cloudflare = {
      source = "cloudflare/cloudflare"
      version = "~> 5"
    }
  }

  required_version = ">= 1.3.0"
}

provider "aws" {
  region = "ap-southeast-1"
}

provider "http" {}

provider "cloudflare" {
  api_token = ""
}