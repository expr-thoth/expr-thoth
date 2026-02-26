package thoth_test

import (
	"context"
	"os"
	"testing"

	thoth "github.com/expr-thoth/expr-thoth"
	"github.com/expr-thoth/expr-thoth/core"
	"github.com/expr-thoth/expr-thoth/internal/testutil"
	"github.com/expr-thoth/expr-thoth/statement"
	"github.com/stretchr/testify/assert"
)

func makePolicyFile(name string) (*statement.Compiler, *thoth.PolicyFile, map[string]any) {
	yamlStr, err := os.ReadFile(name)
	if err != nil {
		panic(err)
	}
	policyFile, err := thoth.NewPolicyFile(string(yamlStr))
	if err != nil {
		panic(err)
	}

	return &statement.Compiler{}, policyFile, thoth.MakeEnv(context.Background())
}

func run(policyFile *thoth.PolicyFile, cmp *statement.Compiler, req *testutil.Request, env map[string]any) *core.Response {
	rsp := &core.Response{}
	_, err := policyFile.Exec(cmp, req, rsp, env)
	if err != nil {
		panic(err)
	}

	return rsp
}

func TestExecBase(t *testing.T) {
	cmp, policyFile, env := makePolicyFile("doc/example/expr.yaml")

	req := &testutil.Request{
		Event:  "user_login_check",
		UserId: "alan turing",
		Attrs:  map[string]string{"role": "admin"},
	}

	run(policyFile, cmp, req, env)
}

func TestForEach(t *testing.T) {
	cmp, policyFile, env := makePolicyFile("doc/example/foreach.yaml")

	req := &testutil.Request{
		Event: "thoth_test_foreach",
	}

	rsp := run(policyFile, cmp, req, env)

	assert.Equal(t, "25", rsp.GetMsg("VarCount"))
}

func TestThenGroup(t *testing.T) {
	cmp, policyFile, env := makePolicyFile("doc/example/thenGroup.yaml")

	req := &testutil.Request{
		Event: "thoth_test_thengroup",
	}

	// 执行过程中输出日志, 应该是乱序的
	rsp := run(policyFile, cmp, req, env)

	// 所有写入操作都应该是成功的
	assert.Equal(t, "10", rsp.GetMsg("A"))
	assert.Equal(t, "11", rsp.GetMsg("B"))
	assert.Equal(t, "12", rsp.GetMsg("C"))
	assert.Equal(t, "13", rsp.GetMsg("D"))
	assert.Equal(t, "14", rsp.GetMsg("E"))
	assert.Equal(t, "15", rsp.GetMsg("F"))
}
