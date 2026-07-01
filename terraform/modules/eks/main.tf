module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 21.0"

  cluster_name    = "gitops-${var.environment}"
  cluster_version = var.cluster_version

  vpc_id     = var.vpc_id
  subnet_ids = var.subnet_ids

  # Allow public API endpoint access (restrict in prod)
  cluster_endpoint_public_access = true

  # Enable IRSA (IAM Roles for Service Accounts)
  enable_irsa = true

  eks_managed_node_groups = {
    default = {
      instance_types = var.instance_types
      min_size       = var.min_size
      max_size       = var.max_size
      desired_size   = var.desired_size

      labels = {
        Environment = var.environment
      }
    }
  }

  # Cluster addons
  cluster_addons = {
    coredns                = {}
    kube-proxy             = {}
    vpc-cni                = {}
    aws-ebs-csi-driver     = {}
  }

  tags = {
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}
