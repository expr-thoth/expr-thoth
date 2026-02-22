package thoth_test

import (
	"context"
	"testing"

	thoth "github.com/expr-thoth/expr-thoth"
	"github.com/expr-thoth/expr-thoth/core"
	"github.com/expr-thoth/expr-thoth/internal/testutil"
	"github.com/expr-thoth/expr-thoth/router"
	"github.com/stretchr/testify/assert"
)

func MakeGroupTestConfig() []router.RouteConfig {
	rst := make([]router.RouteConfig, 0)

	config := router.RouteConfig{
		ActionList: []string{"event_group_test"},
		GroupList:  []string{"testGroup"},
		Enabled:    true,
		Priority:   100,
	}

	rst = append(rst, config)

	return rst
}

func MakeBaseGroupThoth() *thoth.Thoth {
	ctx := context.Background()
	thoth := thoth.NewThoth(ctx)
	thoth.SetBeforePolicy("beforeGroup")
	thoth.SetAfterPolicy("afterGroup")

	thoth.RegisterRouteMap(MakeGroupTestConfig())

	befor := testutil.LoadExampleTestPolicyFile("doc/example/group/before.yaml")
	after := testutil.LoadExampleTestPolicyFile("doc/example/group/after.yaml")
	expr1 := testutil.LoadExampleTestPolicyFile("doc/example/group/expr1.yaml")
	expr2 := testutil.LoadExampleTestPolicyFile("doc/example/group/expr2.yaml")

	thoth.SetGroup("beforeGroup", []*core.PolicyFile{befor})
	thoth.SetGroup("afterGroup", []*core.PolicyFile{after})
	// 反向填充策略文件, 执行时应该按照优先级顺序执行
	thoth.SetGroup("testGroup", []*core.PolicyFile{expr2, expr1})

	return thoth
}

func TestThoth(t *testing.T) {
	thoth := MakeBaseGroupThoth()

	req := &testutil.Request{
		Event:  "event_group_test",
		UserId: "alan turing",
		Attrs:  map[string]string{"role": "admin"},
	}

	rsp, err := thoth.Execute(req)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

	assert.Equal(t, "1", rsp.GetMsg("before_policy_not_halt"))
	assert.Equal(t, "welcome admin", rsp.GetMsg("policy_result"))
	assert.Equal(t, "1", rsp.GetMsg("policy_result2"))
	assert.Equal(t, "done", rsp.GetMsg("after_policy"))

}

func BenchmarkThothBase(b *testing.B) {
	thoth := MakeBaseGroupThoth()

	req := &testutil.Request{
		Event:  "event_group_test",
		UserId: "alan turing",
		Attrs:  map[string]string{"role": "admin"},
	}

	
	for b.Loop() {
		thoth.Execute(req)
	}
}

func TestHalt(t *testing.T) {
	thoth := MakeBaseGroupThoth()

	req := &testutil.Request{
		Event:  "event_group_test",
		UserId: "alan turing",
		Attrs:  map[string]string{"role": "admin", "ControlType": "halt"},
	}

	rsp, err := thoth.Execute(req)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

	assert.Equal(t, "", rsp.GetMsg("before_policy_not_halt"))
	assert.NotEqual(t, "welcome admin", rsp.GetMsg("policy_result"))
	assert.NotEqual(t, "1", rsp.GetMsg("policy_result2"))
	assert.Equal(t, "done", rsp.GetMsg("after_policy"))
}

func TestExitGroup(t *testing.T) {
	thoth := MakeBaseGroupThoth()

	req := &testutil.Request{
		Event: "event_group_test",
		Attrs: map[string]string{"role": "admin", "ControlType": "exit_group"},
	}

	rsp, err := thoth.Execute(req)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

	assert.Equal(t, "1", rsp.GetMsg("before_policy_not_halt"))
	assert.Equal(t, "access denied", rsp.GetMsg("policy_result"))
	assert.NotEqual(t, "1", rsp.GetMsg("policy_result2"))
	assert.Equal(t, "done", rsp.GetMsg("after_policy"))
}

func MakeGrayGroupThoth() *thoth.Thoth {
	ctx := context.Background()
	thoth := thoth.NewThoth(ctx)

	thoth.RegisterRouteMap(MakeGroupTestConfig())
	disable := testutil.LoadExampleTestPolicyFile("doc/example/gray/disable.yaml")
	gray := testutil.LoadExampleTestPolicyFile("doc/example/gray/gray.yaml")

	thoth.SetGroup("testGroup", []*core.PolicyFile{disable, gray})

	return thoth
}

func TestDisable(t *testing.T) {
	thoth := MakeGrayGroupThoth()

	req := &testutil.Request{
		Event:     "event_group_test",
		GrayCount: 0,
		Attrs:     map[string]string{"role": "admin", "ControlType": "exit_group"},
	}

	rsp, err := thoth.Execute(req)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

	assert.NotEqual(t, "1", rsp.GetMsg("disable_policy"))
	assert.Equal(t, "1", rsp.GetMsg("gray_policy"))
}

func TestGray(t *testing.T) {
	thoth := MakeGrayGroupThoth()

	req := &testutil.Request{
		Event:     "event_group_test",
		GrayCount: 1000, // 灰度值大于策略文件指定的值
		Attrs:     map[string]string{"role": "admin", "ControlType": "exit_group"},
	}

	rsp, err := thoth.Execute(req)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

	assert.NotEqual(t, "1", rsp.GetMsg("disable_policy"))
	assert.NotEqual(t, "1", rsp.GetMsg("gray_policy"))
}
