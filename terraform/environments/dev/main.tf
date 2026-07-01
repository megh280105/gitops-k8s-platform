terraform {
  # Uncomment and configure for real AWS deployment with remote state:
  # backend "s3" {
  #   bucket         = "megh-tfstate"
  #   key            = "gitops-k8s/dev/terraform.tfstate"
  #   region         = "us-east-1"
  #   dynamodb_table = "terraform-locks"
  #   encrypt        = true
  # }

  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

provider "aws" {
  region = var.region
}

module "vpc" {
  source      = "../../modules/vpc"
  environment = var.environment
  cidr_block  = var.vpc_cidr
  region      = var.region
}

module "eks" {
  source          = "../../modules/eks"
  environment     = var.environment
  vpc_id          = module.vpc.vpc_id
  subnet_ids      = module.vpc.private_subnet_ids
  cluster_version = "1.31"
  instance_types  = ["t3.medium"]
  min_size        = 2
  max_size        = 4
  desired_size    = 2
}
