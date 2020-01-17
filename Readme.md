# Terratest samples for Azure
[Terratest](https://github.com/gruntwork-io/terratest) is a great tool to test infrastructure. However, the basic sample doesn't seem work on Azure environment. 
As a begineer, it might be painful if you can't run `hello world` stuff. 

I create this repo for my self learning. It might be helpful for terratest begineer for Azure. 

For the Cloud agnostic terratest Getting Started, You can refer: 

* [Terratest](https://github.com/gruntwork-io/terratest) includes Getting Started on top. 

This repo introduces several samples: 

* [terraform AKS deployment](#terraform-AKS-deployment) 
* kubernetes basic
* kubernetes rbac
* helm 

For the kubernetes basic, rbac, and helm are based on the original repo's sample. The original sample seems using NodePort with AWS enviornment. I switch the sample to LoadBalancer and change the code. 

# Prerequisite for the sample 

* [terraform](https://www.terraform.io/downloads.html) is installed and on the PATH environment. 
* [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) is installed on the PATH environment. For kuberntes samples. 
* [helm](https://helm.sh/docs/intro/install/) and [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
* [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest)
* [Go](https://golang.org/doc/install)

I tested both Linux and Windows, however, it might on Mac as well. 

# terraform AKS deployment

## Login using Azure CLI. 

```bash
$ az login
```

In case you are using Azure CLI on WSL

```bash
$ az login --use-device
```

## Create a service principal 

```bash 
$ az ad sp create-for-rbac
```

## Configure Environment Variables

It will be refered from terraform. [TF_VAR_name](https://www.terraform.io/docs/commands/environment-variables.html) used to set variables for terraform. The ARM_SUBSCRIPTION_ID is required for the test. 

```bash
$ export ARM_SUBSCRIPTION_ID={YOUR_SUBSCRIPTION_ID}
$ export TF_VAR_client_id={YOUR_SERVICE_PRINICIPAL_APP_ID} 
$ export TF_VAR_client_secret={YOUR_SERVICE_PRINICIPAL_SECRET}
```

It also requires `~/.ssh/id_rsa.pub` as public key. If you don't have it on that directory, modify the `variables.tf` to point your public key. 

## Create a storage account 
This is an example of using bash. However, you can do similar things on windows, or using Azure Portal. Just create a storage account and container for saving state of terraform. Change the Resource Group. The container name should be `aksstate`. 

```bash
$ RESOURCE_GROUP=MyResourceGroup
$ STORAGE_ACCOUNT_NAME=YOUR_STORAGE_ACCOUNT_NAME
$ CONTAINER_NAME=aksstate
```

```bash
$ az group create -n $RESOURCE_GROUP -l westus
$ az storage account create -n $STORAGE_ACCOUNT_NAME -g $RESOURCE_GROUP -l westus
$ AZURE_STORAGE_KEY=$(az.cmd storage account keys list --account-name tsushistorageaccount --resource-group MyResourceGroup | jq -r .[0].value)
$ az storage container create --name $CONTAINER_NAME --account-name $STORAGE_ACCOUNT_NAME --account-key "$AZURE_STORAGE_KEY"

## Edit Backend config

Go to `test/aks_test.go` then edit this part to fit your resource group and storage account name. Just chenge `storage_account_name` and `resource_group_name`. 

```go
		BackendConfig: map[string]interface{}{
			"storage_account_name": "tsushistatetf",
			"resource_group_name":  "RemoveTerraform",
			"container_name":       "aksstate",
			"key":                  "codelab.microsoft.tfstate",
		},
```
## Run test

initialize the [go module](https://blog.golang.org/using-go-modules), then run the test.  

```bash
$ cd test
$ go mod init github.com/TsuyoshiUshio/TerratestSpikeForAzure
$ go test -v -timeout 30m aks_test.go test_helper.go
```

This will deploy an AKS cluster that defined under the `cluster` directory using terraform. And execute the test if the cluster is up and running.  

If you want to remove the cluster after the test, enable this line on `aks_test.go`

```go
// defer terraform.Destroy(t, terraformOptions)
```

# kubernetes basic 

This sample deploy the nginx by the yaml file under `kubernetes_basic` directory. You ndeed AKS cluster in advance. and put the kubeconfig file under `cluster` directory. If you don't have an AKS cluster, execute `aks_test.go`. You can refer [terraform AKS deployment](#terraform-AKS-deployment) steps. That script create kubeconfig for you under the `cluster` dir. `kubeconfig` is gitignored.  

If you already has a AKS cluster and `culster/kubeconfig` file, run test. 

```bash
$ go test -v -timeout 30m kubernetes_basic_test.go test_helper.go
```

# kubernetse rbac

This sample requires AKS cluster with RBAC enabled. the target yaml file is under `kubernetes_rbac`. It requires AKS cluster and `cluster_rbac/kubeconfig` file

```bash
$ go test -v timeout 30m kubernetes_rbac_test.go test_helper.go
```

If you want to deploy `cluster_rbac` you can follow the instruction [terraform AKS deployment](#terraform-AKS-deployment) and instead of run the test, run terraform by your self.

```bash
$ cd cluster_rbac
(modify cluster_rbac/variables.tf for changing name of cluster)
$ terrafrom apply 
```

# helm

This sample is deploy helm chart(v2). It requires AKS cluster with [tiller](https://v2.helm.sh/docs/using_helm/) installded. `cluster_rbac` script automatically install tiller for you. 

```bash
$ go test -v -timeout 30m helm_test.go test_helper.go
```


