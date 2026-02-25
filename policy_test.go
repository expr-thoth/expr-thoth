package thoth_test

import (
	"context"
	"os"
	"testing"

	"github.com/expr-thoth/expr-thoth"
	"github.com/expr-thoth/expr-thoth/core"
	"github.com/expr-thoth/expr-thoth/internal/testutil"
	"github.com/expr-thoth/expr-thoth/statement"
)

func TestExecBase(t *testing.T) {
	ctx := context.Background()
	cmp := &statement.Compiler{}

	yamlStr, err := os.ReadFile("doc/example/expr.yaml")
	if err != nil {
		t.Fatalf("Failed to read YAML file: %v", err)
	}
	policyFile, err := thoth.NewPolicyFile(string(yamlStr))
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	req := &testutil.Request{
		Event:  "user_login_check",
		UserId: "alan turing",
		Attrs:  map[string]string{"role": "admin"},
	}

	rsp := &core.Response{}
	_, err = policyFile.Exec(cmp, req, rsp, thoth.MakeEnv(ctx))
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

}
