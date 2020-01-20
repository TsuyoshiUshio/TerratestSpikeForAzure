# AKS RBAC enabled cluster configration terraform script

This cluster script is for run rbac related sample. 

## Prerequisite

* Create a Storage Account for terraform state with private container on blob storage.
* install kubectl, helm, and terraform
* Azure CLI

## Deploy 

### Create a service principal for AKS cluster

```bash
$ az login
$ az ad sp create-for-rbac
```

### Set Environment Variables
These commands are for bash example. Just create the enviornment varialbes on Windows. 

```bash
$ export TF_VAR_client_id=YOUR_SERVICE_PRINCIPAL_APP_ID
$ export TF_VAR_client_secret=YOUR_SERVCIE_PRINCIPAL_PASSWORD
```

### Init terraform 
Change the config to fit your Storage Account. You need to modify `tsushistatetf` and `RemoveTerraform`. 

```bash
$ cd cluster_rbac
$ terraform init -backend-config="storage_account_name=tsushistatetf" -backend-config="resource_group_name=RemoveTerraform" -backend-config="container_name=aksstate" -backend-config="key=rbaclab.microsoft.tfstate"
```

### Run terraform 
Change the name of `k8stsushi` as your cluster name. `azure-k8stsushi` as your resource group name. This create a cluster.
This will create kubeconfig file on the current directory, ClusterRoleBindings for default user(it requires for helm), helm resource for installing helm tiller. 

```bash
$ terraform apply -var 'cluster_name=k8stsushi' -var 'resource_group_name=azure-k8stsushi' -var 'location=westus'
```

### Trouble shooting
If the helm installation didn't work, You can do this for installing helm tiller. 
For installing tiller on RBAC configured resource as default, you need to add Cluster Role Bindings. 

```bash
$ az aks get-credentials -n YOUR_CLUSTER_NAME -g YOUR_RESOURCE_GROUP_NAME
$ kubectl apply -f rbac_config.yml
$ helm init
```

Currently these two sample tests are for RBAC
* kubernetes_ingress_integartion_test.go
* kubernetes_rbac_test.go

Happy terratesting!