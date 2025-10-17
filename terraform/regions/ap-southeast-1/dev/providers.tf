terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "~> 6.0"
    }

    cloudflare = {
      source = "cloudflare/cloudflare"
      version = "~> 5"
    }
  }
}

provider "aws" {
  region = "ap-southeast-1"
}

provider "cloudflare" {
  api_token = ""
}