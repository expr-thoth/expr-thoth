package file_test

import (
	"context"
	"os"
	"testing"

	"github.com/expr-thoth/expr-thoth/core"
	"github.com/expr-thoth/expr-thoth/file"
	"github.com/expr-thoth/expr-thoth/internal/testutil"
	"github.com/expr-thoth/expr-thoth/statement"
)

func TestExecBase(t *testing.T) {
	ctx := context.Background()
	cmp := statement.NewCompiler(ctx)
	executor := file.NewExecutor(ctx, cmp)

	yamlStr, err := os.ReadFile("../doc/example/expr.yaml")
	if err != nil {
		t.Fatalf("Failed to read YAML file: %v", err)
	}
	policyFile, err := core.NewPolicyFile(string(yamlStr))
	if err != nil {
		t.Fatalf("Failed to load YAML: %v", err)
	}

	req := &testutil.Request{
		Event:  "user_login_check",
		UserId: "alan turing",
		Attrs:  map[string]string{"role": "admin"},
	}

	rsp := &core.Response{}

	err = executor.Exec(policyFile, req, rsp)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

}
