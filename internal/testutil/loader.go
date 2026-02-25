package testutil

import (
	"os"

	thoth "github.com/expr-thoth/expr-thoth"
)

func LoadExampleTestPolicyFile(name string) *thoth.PolicyFile {
	yamlStr, err := os.ReadFile(name)
	if err != nil {
		panic("Failed to read YAML file: " + err.Error())
	}
	policyFile, err := thoth.NewPolicyFile(string(yamlStr))
	if err != nil {
		panic("Failed to load YAML: " + err.Error())
	}

	return policyFile
}
