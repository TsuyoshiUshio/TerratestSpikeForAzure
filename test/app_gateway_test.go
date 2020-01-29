package test

import (
	"crypto/tls"
	"fmt"
	"testing"
	"time"

	http_helper "github.com/gruntwork-io/terratest/modules/http-helper"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/require"
)

func TestAppGatewayExmple(t *testing.T) {
	t.Parallel()
	expectedResourceGroupName := fmt.Sprintf("tsushi-rg-%s", random.UniqueId())
	expectedApplicationGatewayName := fmt.Sprintf("tsushi-gw-%s", random.UniqueId()) // 1-80 Alphanumeric, hyphen, underscore, and period, case insensitive
	expectedPublicIPAddressName := fmt.Sprintf("%s-pip", expectedApplicationGatewayName)

	terraformOptions := &terraform.Options{
		TerraformDir: "../app_gateway",
		Vars: map[string]interface{}{
			"resource_group_name":      expectedResourceGroupName,
			"application_gateway_name": expectedApplicationGatewayName,
		},
	}
	defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	// Unit Test
	// Wait until the IPAddress is assigned
	WaitUntilPublicIPsAvailable(t, expectedResourceGroupName, expectedPublicIPAddressName, 10, 60*time.Second)

	// Write Application Gateway provisioning state testing

	// IntegrationTest (Http Request for the endpoint)
	publicIPAddress, err := GetPublicIPAddress(expectedResourceGroupName, expectedPublicIPAddressName)
	endpoint := publicIPAddress.PublicIPAddressPropertiesFormat.IPAddress
	require.NoError(t, err)
	tlsConfig := tls.Config{}

	http_helper.HttpGetWithRetryWithCustomValidation(
		t,
		fmt.Sprintf("http://%s", *endpoint),
		&tlsConfig,
		30,
		10*time.Second,
		func(statusCode int, body string) bool {
			return statusCode == 200
		},
	)

}
