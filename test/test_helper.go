package test

import (
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2019-11-01/containerservice"

	"context"
	"fmt"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"
	"os"
	"testing"
	"time"
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
	client, err := GetManagedClustersClient(os.Getenv("ARM_SUBSCRIPTION_ID"))
	if err != nil {
		return nil, err
	}
	managedCluster, err := client.Get(context.Background(), resourceGroupName, clusterName)
	if err != nil {
		return nil, err
	}
	return &managedCluster, nil
}

// WaitUntilServiceExternalIPsAvailable is waiting for allocation of External IP Address
func WaitUntilServiceExternalIPsAvailable(t *testing.T, options *k8s.KubectlOptions, serviceName string, retries int, sleepBetweenRetries time.Duration) {
	statusMsg := fmt.Sprintf("Wait for service %s to be provisioned.", serviceName)
	message := retry.DoWithRetry(
		t,
		statusMsg,
		retries,
		sleepBetweenRetries,
		func() (string, error) {
			service, err := k8s.GetServiceE(t, options, serviceName)
			if err != nil {
				return "", err
			}
			if len(service.Status.LoadBalancer.Ingress) == 0 {
				return "", k8s.NewServiceNotAvailableError(service)
			}
			return "Service ExternalIP is now available", nil
		},
	)
	logger.Logf(t, message)
}
