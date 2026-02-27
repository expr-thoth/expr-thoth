package thoth_test

import (
	"context"
	"testing"

	thoth "github.com/expr-thoth/expr-thoth"
	"github.com/expr-thoth/expr-thoth/internal/testutil"
	"github.com/expr-thoth/expr-thoth/router"
	"github.com/stretchr/testify/assert"
)

func MakeGroupTestConfig() []router.RouteConfig {
	rst := make([]router.RouteConfig, 0)

	config := router.RouteConfig{
		EventList: []string{"event_group_test"},
		GroupList: []string{"testGroup"},
		Enabled:   true,
		Priority:  100,
	}

	rst = append(rst, config)

	return rst
}

func MakeBaseGroupThoth() *thoth.Thoth {
	thothIns := thoth.NewThoth(thoth.WithBeforePolicy("beforeGroup"), thoth.WithAfterPolicy("afterGroup"))

	thothIns.RegisterRouteMap(MakeGroupTestConfig())

	befor := testutil.LoadExampleTestPolicyFile("doc/example/group/before.yaml")
	after := testutil.LoadExampleTestPolicyFile("doc/example/group/after.yaml")
	expr1 := testutil.LoadExampleTestPolicyFile("doc/example/group/expr1.yaml")
	expr2 := testutil.LoadExampleTestPolicyFile("doc/example/group/expr2.yaml")

	thothIns.SetGroup("beforeGroup", []*thoth.PolicyFile{befor})
	thothIns.SetGroup("afterGroup", []*thoth.PolicyFile{after})
	// 反向填充策略文件, 执行时应该按照优先级顺序执行
	thothIns.SetGroup("testGroup", []*thoth.PolicyFile{expr2, expr1})

	return thothIns
}

func TestThoth(t *testing.T) {
	ctx := context.Background()
	thoth := MakeBaseGroupThoth()

	req := &testutil.Request{
		Event:  "event_group_test",
		UserId: "alan turing",
		Attrs:  map[string]string{"role": "admin"},
	}

	rsp, err := thoth.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

	assert.Equal(t, "1", rsp.GetMsg("before_policy_not_halt"))
	assert.Equal(t, "welcome admin", rsp.GetMsg("policy_result"))
	assert.Equal(t, "1", rsp.GetMsg("policy_result2"))
	assert.Equal(t, "done", rsp.GetMsg("after_policy"))

}

func TestHalt(t *testing.T) {
	ctx := context.Background()
	thoth := MakeBaseGroupThoth()

	req := &testutil.Request{
		Event:  "event_group_test",
		UserId: "alan turing",
		Attrs:  map[string]string{"role": "admin", "ControlType": "halt"},
	}

	rsp, err := thoth.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

	assert.Equal(t, "", rsp.GetMsg("before_policy_not_halt"))
	assert.NotEqual(t, "welcome admin", rsp.GetMsg("policy_result"))
	assert.NotEqual(t, "1", rsp.GetMsg("policy_result2"))
	assert.Equal(t, "done", rsp.GetMsg("after_policy"))
}

func TestExitGroup(t *testing.T) {
	ctx := context.Background()
	thoth := MakeBaseGroupThoth()

	req := &testutil.Request{
		Event: "event_group_test",
		Attrs: map[string]string{"role": "admin", "ControlType": "exit_group"},
	}

	rsp, err := thoth.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

	assert.Equal(t, "1", rsp.GetMsg("before_policy_not_halt"))
	assert.Equal(t, "access denied", rsp.GetMsg("policy_result"))
	assert.NotEqual(t, "1", rsp.GetMsg("policy_result2"))
	assert.Equal(t, "done", rsp.GetMsg("after_policy"))
}

func MakeGrayGroupThoth() *thoth.Thoth {
	thothIns := thoth.NewThoth()

	thothIns.RegisterRouteMap(MakeGroupTestConfig())
	disable := testutil.LoadExampleTestPolicyFile("doc/example/gray/disable.yaml")
	gray := testutil.LoadExampleTestPolicyFile("doc/example/gray/gray.yaml")

	thothIns.SetGroup("testGroup", []*thoth.PolicyFile{disable, gray})

	return thothIns
}

func TestDisable(t *testing.T) {
	ctx := context.Background()
	thoth := MakeGrayGroupThoth()

	req := &testutil.Request{
		Event:     "event_group_test",
		GrayCount: 0,
		Attrs:     map[string]string{"role": "admin", "ControlType": "exit_group"},
	}

	rsp, err := thoth.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

	assert.NotEqual(t, "1", rsp.GetMsg("disable_policy"))
	assert.Equal(t, "1", rsp.GetMsg("gray_policy"))
}

func TestGray(t *testing.T) {
	ctx := context.Background()
	thoth := MakeGrayGroupThoth()

	req := &testutil.Request{
		Event:     "event_group_test",
		GrayCount: 1000, // 灰度值大于策略文件指定的值
		Attrs:     map[string]string{"role": "admin", "ControlType": "exit_group"},
	}

	rsp, err := thoth.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

	assert.NotEqual(t, "1", rsp.GetMsg("disable_policy"))
	assert.NotEqual(t, "1", rsp.GetMsg("gray_policy"))
}

func MakeFlowGroupThoth() *thoth.Thoth {
	thothIns := thoth.NewThoth(thoth.WithBeforePolicy("beforeFlow"), thoth.WithAfterPolicy("afterFlow"))

	// 注册路由配置
	routeConfig := router.RouteConfig{
		EventList: []string{"event_flow_test"},
		GroupList: []string{"mainFlow"},
		Enabled:   true,
		Priority:  100,
	}
	thothIns.RegisterRouteMap([]router.RouteConfig{routeConfig})

	// 加载 flow 相关的策略文件
	before1 := testutil.LoadExampleTestPolicyFile("doc/example/flow/before/01.yaml")
	before2 := testutil.LoadExampleTestPolicyFile("doc/example/flow/before/02.yaml")
	after := testutil.LoadExampleTestPolicyFile("doc/example/flow/after.yaml")
	main1 := testutil.LoadExampleTestPolicyFile("doc/example/flow/main/01.yaml")
	main2 := testutil.LoadExampleTestPolicyFile("doc/example/flow/main/02.yaml")

	// 设置策略组
	thothIns.SetGroup("beforeFlow", []*thoth.PolicyFile{before1, before2})
	thothIns.SetGroup("afterFlow", []*thoth.PolicyFile{after})
	thothIns.SetGroup("mainFlow", []*thoth.PolicyFile{main1, main2})

	return thothIns
}

func TestFlowControl(t *testing.T) {
	ctx := context.Background()
	thothIns := MakeFlowGroupThoth()

	req := &testutil.Request{
		Event: "event_flow_test",
		Attrs: map[string]string{"ControlType": "normal"},
	}

	rsp, err := thothIns.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

	// 测试正常执行情况
	assert.Equal(t, "1", rsp.GetMsg("befeore_policy_02"))
	assert.Equal(t, "1", rsp.GetMsg("main_policy_02"))
}

func TestFlowHalt(t *testing.T) {
	ctx := context.Background()
	thothIns := MakeFlowGroupThoth()

	req := &testutil.Request{
		Event: "event_flow_test",
		Attrs: map[string]string{"ControlType": "halt"},
	}

	rsp, err := thothIns.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

	// 测试 Halt 功能
	assert.Equal(t, "1", rsp.GetMsg("before_halt"))
	assert.Empty(t, rsp.GetMsg("after_halt"))
	assert.Empty(t, rsp.GetMsg("after_halt_other_action"))
	assert.Empty(t, rsp.GetMsg("befeore_policy_02"))
	assert.Empty(t, rsp.GetMsg("main_policy_02"))
}

func TestFlowExitGroup(t *testing.T) {
	ctx := context.Background()
	thothIns := MakeFlowGroupThoth()

	req := &testutil.Request{
		Event: "event_flow_test",
		Attrs: map[string]string{"ControlType": "exit_group"},
	}

	rsp, err := thothIns.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

	// 测试 ExitGroup 功能
	assert.Equal(t, "1", rsp.GetMsg("exit_group_thenlist"))
	assert.Empty(t, rsp.GetMsg("after_exit_group_thenlist"))
	assert.Empty(t, rsp.GetMsg("after_exit_group_action"))
	assert.Empty(t, rsp.GetMsg("main_policy_02"))
}

func TestFlowReturn(t *testing.T) {
	ctx := context.Background()
	thothIns := MakeFlowGroupThoth()

	req := &testutil.Request{
		Event: "event_group_test",
	}

	rsp, err := thothIns.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Failed to execute policy: %v", err)
	}

	// 测试 Return 功能
	assert.Equal(t, "1", rsp.GetMsg("after_policy"))
	assert.Equal(t, "1", rsp.GetMsg("before_return_thenlist"))
	assert.Empty(t, rsp.GetMsg("after_return_thenlist"))
	assert.Empty(t, rsp.GetMsg("after_return_action"))
}
