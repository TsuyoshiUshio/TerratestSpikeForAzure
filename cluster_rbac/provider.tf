provider "azurerm" {
    version = "1.40.0"
}

terraform {
    backend "azurerm" {}
}

provider "kubernetes" {
    config_path = "./kubeconfig"
}

provider "helm" {
  kubernetes {
      config_path = "./kubeconfig"
  }
}