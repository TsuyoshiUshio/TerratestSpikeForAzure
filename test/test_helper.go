package test

import (
	"io/ioutil"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2019-11-01/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-09-01/network"

	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/retry"

	gwErrors "github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/gruntwork-cli/files"
)

const (
	SubscriptionIDEnv = "ARM_SUBSCRIPTION_ID"
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

// GetApplicationGatewayClient creates a client
func GetApplicationGatewayClient(subscriptionID string) (*network.ApplicationGatewaysClient, error) {
	client := network.NewApplicationGatewaysClient(subscriptionID)
	authorizer, err := NewAuthorizer()
	if err != nil {
		return nil, err
	}
	client.Authorizer = *authorizer
	return &client, nil
}

// GetPublicIPAddressClient creates a client
func GetPublicIPAddressClient(subscriptionID string) (*network.PublicIPAddressesClient, error) {
	client := network.NewPublicIPAddressesClient(subscriptionID)
	authorizer, err := NewAuthorizer()
	if err != nil {
		return nil, err
	}
	client.Authorizer = *authorizer
	return &client, nil
}

// NewAuthorizer will return Authorizer
func NewAuthorizer() (*autorest.Authorizer, error) {
	authorizer, err := auth.NewAuthorizerFromCLI()
	return &authorizer, err
}

// GetManagedCluster will return ContainerService
func GetManagedCluster(resourceGroupName, clusterName string) (*containerservice.ManagedCluster, error) {
	client, err := GetManagedClustersClient(os.Getenv(SubscriptionIDEnv))
	if err != nil {
		return nil, err
	}
	managedCluster, err := client.Get(context.Background(), resourceGroupName, clusterName)
	if err != nil {
		return nil, err
	}
	return &managedCluster, nil
}

// GetApplicationGateway will return ApplicationGatway
func GetApplicationGateway(resourceGroupName, applicationGatewayName string) (*network.ApplicationGateway, error) {
	client, err := GetApplicationGatewayClient(os.Getenv(SubscriptionIDEnv))
	if err != nil {
		return nil, err
	}
	applicationGateway, err := client.Get(context.Background(), resourceGroupName, applicationGatewayName)
	if err != nil {
		return nil, err
	}
	return &applicationGateway, nil
}

// GetPublicIPAddress will return PublicIPAddress
func GetPublicIPAddress(resourceGroupName, publicIPAddressName string) (*network.PublicIPAddress, error) {
	client, err := GetPublicIPAddressClient(os.Getenv(SubscriptionIDEnv))
	if err != nil {
		return nil, err
	}
	publicIPAddress, err := client.Get(context.Background(), resourceGroupName, publicIPAddressName, "")
	if err != nil {
		return nil, err
	}
	return &publicIPAddress, nil
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

// WaitUntilPublicIPsAvailable is waiting for allocation of External IP Address
func WaitUntilPublicIPsAvailable(t *testing.T, resourceGroupName, publicIPAddressName string, retries int, sleepBetweenRetries time.Duration) {
	statusMsg := fmt.Sprintf("Wait for PublicIPAddress %s to be provisioned.", publicIPAddressName)
	message := retry.DoWithRetry(
		t,
		statusMsg,
		retries,
		sleepBetweenRetries,
		func() (string, error) {
			publicIPAddress, err := GetPublicIPAddress(resourceGroupName, publicIPAddressName)
			if err != nil {
				return "", err
			}

			if publicIPAddress.PublicIPAddressPropertiesFormat.IPAddress == nil {
				return "", fmt.Errorf("PublicIPAddress %s has not assigned yet", publicIPAddressName)
			}
			return "Service ExternalIP is now available", nil
		},
	)
	logger.Logf(t, message)
}

func copyKubeConfigToTempE(t *testing.T, configPath string) (string, error) {
	tmpConfig, err := ioutil.TempFile("", "")
	if err != nil {
		return "", gwErrors.WithStackTrace(err)
	}
	defer tmpConfig.Close()
	err = files.CopyFile(configPath, tmpConfig.Name())
	return tmpConfig.Name(), err
}
