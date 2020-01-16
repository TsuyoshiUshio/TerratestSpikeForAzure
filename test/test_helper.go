package test

import (
  "github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2019-11-01/containerservice"

  "github.com/Azure/go-autorest/autorest"
  "github.com/Azure/go-autorest/autorest/azure/auth"
  "context"
  "os"
)

// GetManagedClustersClient creates a client
func GetManagedClustersClient(subscriptionID string) (*containerservice.ManagedClustersClient, error) {
	managedServicesClient := containerservice.NewManagedClustersClient(subscriptionID)
	authorizer, err := NewAuthorizer()

	if err != nil {
		return nil, err
	}

	managedServicesClient.Authorizer = *authorizer

	return &managedServicesClient, nil

}

// NewAuthorizer will return Authorizer
func NewAuthorizer() (*autorest.Authorizer, error) {
	authorizer, err := auth.NewAuthorizerFromCLI()
	return &authorizer, err
}

// GetManagedCluster will return ContainerService
func GetManagedCluster(resourceGroupName, clusterName string) (*containerservice.ManagedCluster, error) {
	client, err := GetManagedClustersClient(os.Getenv("SUBSCRIPTION_ID"))
	if err != nil {
		return nil, err
	}
    managedCluster, err :=	client.Get(context.Background(), resourceGroupName, clusterName)
	if err != nil {
		return nil, err
	}
	return &managedCluster, nil
}