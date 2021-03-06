resource "azurerm_resource_group" "k8s" {
    name     = var.resource_group_name
    location = var.location
}

resource "azurerm_kubernetes_cluster" "k8s" {
    name                = var.cluster_name
    location            = azurerm_resource_group.k8s.location
    resource_group_name = azurerm_resource_group.k8s.name
    dns_prefix          = var.dns_prefix
    role_based_access_control {
        enabled = true
    }

    linux_profile {
        admin_username = "ubuntu"

        ssh_key {
            key_data = file(var.ssh_public_key)
        }
    }

    default_node_pool {
        name            = "agentpool"
        node_count      = var.agent_count
        vm_size         = "Standard_DS1_v2"
    }

    service_principal {
        client_id     = var.client_id
        client_secret = var.client_secret
    }

    tags = {
        Environment = "Development"
    }
}

resource "kubernetes_cluster_role_binding" "rbac" {
  metadata {
    name = "cluster-admin-management-binding"
  }
  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = "cluster-admin"
  }
  subject {
    kind      = "User"
    name      = "system:serviceaccount:kube-system:default"
    api_group = "rbac.authorization.k8s.io"
  }
      depends_on = [
        azurerm_kubernetes_cluster.k8s
    ]
}

resource "helm_release" "wordpress" {
    name = "my-wordpress"
    chart = "stable/wordpress"

    depends_on = [
        kubernetes_cluster_role_binding.rbac,
        local_file.kubeconfig
    ]
}

resource "local_file" "kubeconfig" {
    content     = azurerm_kubernetes_cluster.k8s.kube_config_raw
    filename = "kubeconfig"

    depends_on = [
        azurerm_kubernetes_cluster.k8s
    ]
}
