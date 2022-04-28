package utils

import (
	"fmt"
	"os/exec"
)

func DescribeAPIClarityDeployment() {
	cmd := exec.Command("kubectl", "-n", "apiclarity", "describe", "deployments.apps", APIClarityDeploymentName)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to execute command. %v, %s", err, out)
		return
	}
	fmt.Printf("kubectl describe deployments.apps -n apiclarity apiclarity-apiclarity:\n %s\n", out)
}

func DescribeAPIClarityPods() {
	cmd := exec.Command("kubectl", "-n", "apiclarity", "describe", "pods")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to execute command. %v, %s", err, out)
		return
	}
	fmt.Printf("kubectl describe pods -n apiclarity:\n %s\n", out)
}

