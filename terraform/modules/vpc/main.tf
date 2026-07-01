locals {
  azs             = ["${var.region}a", "${var.region}b", "${var.region}c"]
  public_subnets  = [for i, az in local.azs : cidrsubnet(var.cidr_block, 4, i)]
  private_subnets = [for i, az in local.azs : cidrsubnet(var.cidr_block, 4, i + 3)]
}

module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "gitops-${var.environment}"
  cidr = var.cidr_block
  azs  = local.azs

  public_subnets  = local.public_subnets
  private_subnets = local.private_subnets

  enable_nat_gateway     = true
  single_nat_gateway     = var.environment != "prod"
  enable_dns_hostnames   = true
  enable_dns_support     = true

  # Required tags for EKS subnet auto-discovery
  public_subnet_tags = {
    "kubernetes.io/role/elb" = 1
  }
  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = 1
  }

  tags = {
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}
