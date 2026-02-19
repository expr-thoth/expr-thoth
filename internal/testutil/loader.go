package testutil

import (
	"os"

	"github.com/expr-thoth/expr-thoth/core"
)

func LoadExampleTestPolicyFile(name string) *core.PolicyFile {
	yamlStr, err := os.ReadFile(name)
	if err != nil {
		panic("Failed to read YAML file: " + err.Error())
	}
	policyFile, err := core.NewPolicyFile(string(yamlStr))
	if err != nil {
		panic("Failed to load YAML: " + err.Error())
	}

	return policyFile
}
