variable "environment" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "subnet_ids" {
  type = list(string)
}

variable "cluster_version" {
  type    = string
  default = "1.31"
}

variable "instance_types" {
  type    = list(string)
  default = ["t3.medium"]
}

variable "min_size" {
  type    = number
  default = 2
}

variable "max_size" {
  type    = number
  default = 5
}

variable "desired_size" {
  type    = number
  default = 2
}
