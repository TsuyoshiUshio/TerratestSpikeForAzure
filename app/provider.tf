provider "helm" {
  kubernetes {
      config_path = "../cluster/kubeconfig"
  }
}