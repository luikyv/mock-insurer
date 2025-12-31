variable "subscription_id" {
  description = "Azure subscription ID"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
}

variable "project" {
  description = "Project name"
  type        = string
  default     = "mock-insurer"
}

variable "env" {
  description = "Environment"
  type        = string
  default     = "dev"
}

variable "domain" {
  description = "Domain"
  type        = string
}

variable "org_id" {
  description = "Organization ID"
  type        = string
}

variable "image_tag" {
  description = "Image tag"
  type        = string
  default     = "latest"
}