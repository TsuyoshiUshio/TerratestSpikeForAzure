provider "azurerm" {
    version = "1.40.0"
}

terraform {
    backend "azurerm" {}
}

provider "helm" {
  kubernetes {
      config_path = "./kubeconfig"
  }
}