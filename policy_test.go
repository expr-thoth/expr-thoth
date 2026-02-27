package thoth_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	thoth "github.com/expr-thoth/expr-thoth"
	"github.com/expr-thoth/expr-thoth/core"
	"github.com/expr-thoth/expr-thoth/internal/testutil"
	"github.com/expr-thoth/expr-thoth/internal/testutil/plugins/log"
	plgPanic "github.com/expr-thoth/expr-thoth/internal/testutil/plugins/panic"
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

func run(policyFile *thoth.PolicyFile, cmp *statement.Compiler, req *testutil.Request, env map[string]any) (*core.Response, error) {
	rsp := &core.Response{}
	_, err := policyFile.Exec(cmp, req, rsp, env)

	return rsp, err
}

func TestExecBase(t *testing.T) {
	cmp, policyFile, env := makePolicyFile("doc/example/expr.yaml")

	req := &testutil.Request{
		Event:  "user_login_check",
		UserId: "alan turing",
		Attrs:  map[string]string{"role": "admin"},
	}

	rsp, err := run(policyFile, cmp, req, env)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, 200, rsp.Code)
	assert.Equal(t, "welcome turing", rsp.GetMsg("policy_result"))

	fmt.Print(rsp)
}

func TestForEach(t *testing.T) {
	cmp, policyFile, env := makePolicyFile("doc/example/foreach.yaml")

	req := &testutil.Request{
		Event: "thoth_test_foreach",
	}

	rsp, err := run(policyFile, cmp, req, env)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, "25", rsp.GetMsg("VarCount"))
}

func TestThenGroup(t *testing.T) {
	cmp, policyFile, env := makePolicyFile("doc/example/then_group.yaml")

	req := &testutil.Request{
		Event: "event_test_thengroup",
	}

	// 执行过程中输出日志, 应该是乱序的
	rsp, err := run(policyFile, cmp, req, env)
	if err != nil {
		panic(err)
	}

	// 所有写入操作都应该是成功的
	assert.Equal(t, "10", rsp.GetMsg("A"))
	assert.Equal(t, "11", rsp.GetMsg("B"))
	assert.Equal(t, "12", rsp.GetMsg("C"))
	assert.Equal(t, "13", rsp.GetMsg("D"))
	assert.Equal(t, "14", rsp.GetMsg("E"))
	assert.Equal(t, "15", rsp.GetMsg("F"))
}

func TestThenGroupPainc(t *testing.T) {
	inst := thoth.NewThoth()
	inst.RegisterPlugin("Panic", plgPanic.Create)
	cmp, policyFile, env := makePolicyFile("doc/example/panic_group.yaml")

	req := &testutil.Request{
		Event: "event_test_thengroup",
	}

	_, err := run(policyFile, cmp, req, env)

	// 由于产生了panic, 因此必然有错误日志
	assert.Error(t, err)
	fmt.Println(err)
}

func TestLog(t *testing.T) {
	inst := thoth.NewThoth()
	inst.RegisterPlugin("Log", log.Create)
	cmp, policyFile, env := makePolicyFile("doc/example/log.yaml")

	req := &testutil.Request{
		Event: "event_log_test",
	}

	_, err := run(policyFile, cmp, req, env)
	if err != nil {
		panic(err)
	}
}

func BenchmarkBase(b *testing.B) {
	cmp, policyFile, _ := makePolicyFile("doc/example/bench.yaml")

	req := &testutil.Request{
		Event:  "user_login_check",
		UserId: "alan turing",
		Attrs:  map[string]string{"role": "admin"},
	}

	for b.Loop() {
		env := thoth.MakeEnv(context.Background())
		run(policyFile, cmp, req, env)
	}
}