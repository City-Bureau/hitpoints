variable "resource_group_location" {
  description = "Azure resource group location"
  default     = "East US"
}

variable "tags" {
  description = "Additional tags"
  type        = map
  default     = {}
}

variable "vm_size" {
  description = "VM size"
  default     = "Standard_B1ls"
}

variable "domain" {
  description = "Domain the server should respond to"
}

variable "prefix" {
  default = "hitpoints"
}

variable "ssh_pub_key" {
  description = "SSH public key to be provisioned on the instance"
}

variable "ssh_priv_key_path" {
  description = "Path to private SSH key"
}

variable "ssl" {
  type        = bool
  description = "Whether SSL should be enabled"
  default     = true
}

variable "allow_ssh" {
  type        = bool
  description = "Whether SSH connections should be allowed"
  default     = true
}
