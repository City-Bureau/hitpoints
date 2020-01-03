variable "region" {
  description = "AWS Region"
  default     = "us-east-1"
}

variable "vpc_id" {
  description = "AWS VPC to use"
}

variable "subnet_id" {
  description = "AWS public subnet ID"
}

variable "tags" {
  description = "Additional tags"
  type        = map
  default     = {}
}

variable "instance_type" {
  description = "The type of the instance"
  default     = "t2.nano"
}

variable "domain" {
  description = "Domain the server should respond to"
}

variable "prefix" {
  default = "hitpoints"
}

variable "keyname" {
  description = "SSH key name"
}

variable "ssh_priv_key_path" {
  description = "Path to private SSH key for keyname"
}

variable "ssl" {
  type        = bool
  description = "Whether SSL should be enabled"
  default     = true
}
