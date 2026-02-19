package group_test

import (
	"context"
	"testing"

	"github.com/expr-thoth/expr-thoth/core"
	"github.com/expr-thoth/expr-thoth/group"
	"github.com/expr-thoth/expr-thoth/internal/testutil"
)

func TestGroup(t *testing.T) {
	ctx := context.Background()
	exector := group.NewExecutor(ctx)

	policyFile := testutil.LoadExampleTestPolicyFile("../doc/example/expr.yaml")

	files := []*core.PolicyFile{policyFile}
	exector.UpdateGroup("testGroup", files)

	req := &testutil.Request{
		Event:  "user_login_check",
		UserId: "alan turing",
		Attrs:  map[string]string{"role": "admin"},
	}

	rsp := &core.Response{}
	_, err := exector.Exec("testGroup", req, rsp)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

}
