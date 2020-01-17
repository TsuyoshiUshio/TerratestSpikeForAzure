package test

import (
	"fmt"
	"testing"

	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAKSExample(t *testing.T) {
	t.Parallel()
	deployKubernetes(t)
	// deployHelmChart(t)
}

func deployKubernetes(t *testing.T) {
	expectedClusterName := fmt.Sprintf("tsushi-%s", random.UniqueId()) // UniqueId length is 6.
	expectedResourceGroupName := "RemoveAKS"
	expectedAagentCount := 3

	terraformOptions := &terraform.Options{
		TerraformDir: "../cluster",
		BackendConfig: map[string]interface{}{
			"storage_account_name": "tsushistatetf",
			"resource_group_name":  "RemoveTerraform",
			"container_name":       "aksstate",
			"key":                  "codelab.microsoft.tfstate",
		},
		Vars: map[string]interface{}{
			"cluster_name":        expectedClusterName,
			"resource_group_name": expectedResourceGroupName,
			"agent_count":         expectedAagentCount,
		},
	}

	// defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	// Lookup the output variable
	// some := terraform.Output(t, terraformOptions, "some")
	// kubeconfig := terraform.Output(t, terraformOptions, "kube_config")
	// Look up the AKS cluster
	// method 1. Azure SDK
	// method 2. Azure CLI (external execution)
	var ActualCount int32 = 0
	cluster, err := GetManagedCluster(expectedResourceGroupName, expectedClusterName)
	assert.Nil(t, err)
	ActualCount = *(*cluster.ManagedClusterProperties.AgentPoolProfiles)[0].Count
	// Write assert
	assert.Equal(t, int32(expectedAagentCount), ActualCount)
}

func deployHelmChart(t *testing.T) {
	terraformOptions := &terraform.Options{
		TerraformDir: "../app",
		BackendConfig: map[string]interface{}{
			"storage_account_name": "tsushistatetf",
			"resource_group_name":  "RemoveTerraform",
			"container_name":       "aksstate",
			"key":                  "codeapp.microsoft.tfstate",
		},
	}

	// defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

}
