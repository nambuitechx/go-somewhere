terraform {
  backend "s3" {
    bucket = "go-somewhere-tfstate-832557411742"
    key    = "dev/terraform.tfstate"
    region = "ap-southeast-1"
  }
}