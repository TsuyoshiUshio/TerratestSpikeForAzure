variable "resource_group_name" {
    default = "RemoveAGW"
}

variable "application_gateway_name" {

}

variable "location" {
    default = "West US"
}

variable "prefix" {
  default = "tfvmex"
}

variable "cloudconfig_file" {
  default = "./cloud-init.txt" 
}