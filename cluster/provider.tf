provider "azurerm" {
    version = "1.40.0"
}

terraform {
    backend "azurerm" {}
}