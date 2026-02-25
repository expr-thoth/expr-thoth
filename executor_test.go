package thoth_test

// TODO: 无法构造一个单独的执行器, 因此无法单独测试

// import (
// 	"context"
// 	"testing"

// 	"github.com/expr-thoth/expr-thoth"
// 	"github.com/expr-thoth/expr-thoth/core"
// 	"github.com/expr-thoth/expr-thoth/internal/testutil"
// )

// func TestGroup(t *testing.T) {
// 	ctx := context.Background()
// 	opts := *&thoth.Options{

// 	}

// 	exector := thoth.NewExecutor(ctx, )

// 	policyFile := testutil.LoadExampleTestPolicyFile("doc/example/expr.yaml")

// 	files := []*thoth.PolicyFile{policyFile}
// 	exector.UpdateGroup("testGroup", files)

// 	req := &testutil.Request{
// 		Event:  "user_login_check",
// 		UserId: "alan turing",
// 		Attrs:  map[string]string{"role": "admin"},
// 	}

// 	rsp := &core.Response{}
// 	_, err := exector.Exec("testGroup", req, rsp)
// 	if err != nil {
// 		t.Fatalf("Failed to execute policy: %v", err)
// 	}

// }
