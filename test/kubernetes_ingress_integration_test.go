package test

import (
	"crypto/tls"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gruntwork-io/terratest/modules/helm"
	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
)

// The test for deploying ingress controller
func TestKubernetsIngressIntegrationExampleDeployment(t *testing.T) {
	t.Parallel()

	const serviceAccountName = "terratest-ingress-example-service-account"
	const namespaceName = "terratest-ingress-example-namespace"

	// Path to the RBAC configration
	serviceAccountResourcePath, err := filepath.Abs("../kubernetes_ingress/namespace-service-account.yml")
	require.NoError(t, err)

	// Setup the kubectl config and context. Here we choose to create a new one because we will be manipulating the
	// entries to be able to add a new authentication option.
	tmpConfigPath, err := copyKubeConfigToTempE(t, "../cluster_rbac/kubeconfig")
	require.NoError(t, err)
	defer os.Remove(tmpConfigPath)
	options := k8s.NewKubectlOptions("", tmpConfigPath, namespaceName)

	// At the end of the test, run `kubectl delete -f RESOURCE_CONFIG` to clean up any resources that were created.
	defer k8s.KubectlDelete(t, options, serviceAccountResourcePath)

	serviceAccountKubectlOptions := setupServiceAccount(t, serviceAccountResourcePath, tmpConfigPath, serviceAccountName, namespaceName)

	// Setup the args. For this test, we will set the following input values:
	// - containerImageRepo=nginx
	// - containerImageTag=1.15.8
	helmOptions := &helm.Options{
		KubectlOptions: serviceAccountKubectlOptions,
	}

	chart := "aks-helloworld"
	releaseName := "aks-helloworld"
	defer helm.Delete(t, helmOptions, releaseName, true)
	applyHelm(t, helmOptions, chart, releaseName, "https://azure-samples.github.io/helm-charts")
	k8s.WaitUntilServiceAvailable(t, serviceAccountKubectlOptions, releaseName, 10, 2*time.Second)

	releaseName = "aks-helloworld-two"
	defer helm.Delete(t, helmOptions, releaseName, true)
	applyHelm(t, helmOptions, chart, releaseName, "https://azure-samples.github.io/helm-charts", "--set", "serviceName="+releaseName, "--set", "title=\"AKS Ingress Demo\"")
	k8s.WaitUntilServiceAvailable(t, serviceAccountKubectlOptions, releaseName, 10, 2*time.Second)

	// Deploy nginx-ingess helm chart for sample application.
	// heml.Install doesn't support --repo.
	chart = "nginx-ingress"
	releaseName = fmt.Sprintf(
		"nginx-ingress-%s",
		strings.ToLower(random.UniqueId()),
	)
	defer helm.Delete(t, helmOptions, releaseName, true)

	applyHelm(t, helmOptions, chart, releaseName, "https://kubernetes-charts.storage.googleapis.com/")

	// Path to the helm chart we will test
	appResourcePath, err := filepath.Abs("../kubernetes_ingress/hello-world-ingress.yml")
	require.NoError(t, err)

	defer k8s.KubectlDelete(t, serviceAccountKubectlOptions, appResourcePath)

	// This will run `kubectl apply -f RESOURCE_CONFIG` and fail the test if there are any errors
	k8s.KubectlApply(t, serviceAccountKubectlOptions, appResourcePath)

	// Now let's verify the deployment. We will get the service endpoint and try to access it.
	serviceName := fmt.Sprintf("%s-controller", releaseName)

	// This will wait up to 10 seconds for the service to become available, to ensure that we can access it.
	// k8s.WaitUntilServiceAvailable(t, options, "nginx-service", 10, 1*time.Second)
	WaitUntilServiceExternalIPsAvailable(t, serviceAccountKubectlOptions, serviceName, 10, 20*time.Second)
	// Now we verify that the service will successfully boot and start serving requests
	service := k8s.GetService(t, serviceAccountKubectlOptions, serviceName)
	endpoint := service.Status.LoadBalancer.Ingress[0].IP

	// Setup a TLS configuration to submit with the helper, a blank struct is acceptable
	tlsConfig := tls.Config{}

	// Test the endpoint for up to 5 minutes. This will only fail if we timeout waiting for the service to return a 200
	// response.
	http_helper.HttpGetWithRetryWithCustomValidation(
		t,
		fmt.Sprintf("http://%s", endpoint),
		&tlsConfig,
		30,
		10*time.Second,
		func(statusCode int, body string) bool {
			return statusCode == 200
		},
	)

	http_helper.HttpGetWithRetryWithCustomValidation(
		t,
		fmt.Sprintf("http://%s/hello-world-two", endpoint),
		&tlsConfig,
		30,
		10*time.Second,
		func(statusCode int, body string) bool {
			return statusCode == 200
		},
	)
}

func setupServiceAccount(t *testing.T, serviceAccountResourcePath, tmpConfigPath, serviceAccountName, namespaceName string) *k8s.KubectlOptions {
	options := k8s.NewKubectlOptions("", tmpConfigPath, namespaceName)

	// This will run `kubectl apply -f RESOURCE_CONFIG` and fail the test if there are any errors
	k8s.KubectlApply(t, options, serviceAccountResourcePath)

	// Retrieve authentication token for the newly created ServiceAccount
	token := k8s.GetServiceAccountAuthToken(t, options, serviceAccountName)

	// Now update the configuration to add a new context that can be used to make requests as that service account
	require.NoError(t, k8s.AddConfigContextForServiceAccountE(
		t,
		options,
		serviceAccountName, // for this test we will name the context after the ServiceAccount
		serviceAccountName,
		token,
	))
	return k8s.NewKubectlOptions(serviceAccountName, tmpConfigPath, namespaceName)
}

func applyHelm(t *testing.T, helmOptions *helm.Options, chart, releaseName, repo string, optionalArgs ...string) {
	args := []string{}
	args = append(args, "--namespace", helmOptions.KubectlOptions.Namespace)
	args = append(args, "--set", "controller.replicaCount=2")
	// args = append(args, "--set", "controller.nodeSelector.\"beta.kubernetes.io/os\"=linux")
	// args = append(args, "--set", "defaultBackend.nodeSelector.\"beta.kubernetes.io/os\"=linux")
	args = append(args, optionalArgs...)
	args = append(args, "--repo", repo)
	args = append(args, "-n", releaseName, chart)
	_, err := helm.RunHelmCommandAndGetOutputE(t, helmOptions, "install", args...)
	if err != nil {
		t.Log(err)
	}
}
