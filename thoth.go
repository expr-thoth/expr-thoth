package thoth

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/expr-thoth/expr-thoth/core"
	"github.com/expr-thoth/expr-thoth/group"
	"github.com/expr-thoth/expr-thoth/router"
)

// Thoth是整个策略引擎的核心, 负责管理策略分组, 路由表以及策略执行器等核心组件
type Thoth struct {
	initPolicyGroupName   string
	beforePolicyGroupName string
	afterPolicyGroupName  string
	routeMap              atomic.Pointer[router.RouteMap]
	executor              *group.Executor
}

// NewThoth创建Thoth实例, 需要传入上下文对象, 以便在策略执行过程中进行日志记录等操作
func NewThoth(ctx context.Context) *Thoth {
	return &Thoth{
		executor: group.NewExecutor(ctx),
	}
}

// 注册插件, 可使用Sys变量加载注册的插件
// 初始化阶段调用, 非线程安全
func (t *Thoth) RegisterPlugin(name string, f core.PluginFactory) error {
	core.RegisterPlugin(name, f)
	return nil
}

// 设置Init分组策略, 可在该分组配置仅需单次执行, 与请求无关的初始化相关的逻辑
// 初始化阶段调用, 非线程安全
func (t *Thoth) SetInitPolicy(groupName string) {
	t.initPolicyGroupName = groupName
}

// 执行Init分组策略, 业务决定具体调用时机, 例如启动时执行一次, 或者Init策略发生变更时执行一次
// 是否线程安全取决于业务实现
func (t *Thoth) ExecuteInitPolicy() error {
	if g := t.initPolicyGroupName; g != "" {
		_, err := t.executor.Exec(g, core.InitRequest{}, &core.Response{})
		if err != nil {
			return err
		}
	}
	return nil
}

// 设置前置策略分组名, 执行策略时所有事件始终先执行前置策略
// 初始化阶段调用, 非线程安全
func (t *Thoth) SetBeforePolicy(groupName string) {
	t.beforePolicyGroupName = groupName
}

// 注册策略路由表, 执行前置策略后, 按照路由表决定事件后续执行的分组
// 线程安全函数, 支持热加载路由表
func (t *Thoth) RegisterRouteMap(configs []router.RouteConfig) {
	routeMap := router.New(configs)
	t.routeMap.Store(&routeMap)
}

// 设置后置策略分组名, 其余策略执行完毕后, 最后执行后置策略
// 无论策略中是否发生错误或者执行控制流指令, 只要后置策略配置了, 就无条件执行
// 初始化阶段调用, 非线程安全
func (t *Thoth) SetAfterPolicy(groupName string) {
	t.afterPolicyGroupName = groupName
}

// 设置分组数据, 其存储方式由调用方决定, 可以是本地文件, 也可以是远程配置中心等
// 线程安全函数, 支持热加载分组数据
func (t *Thoth) SetGroup(groupName string, policyFiles []*core.PolicyFile) {
	t.executor.UpdateGroup(groupName, policyFiles)
}

// 策略入口函数, 其执行流程为: 前置策略 -> 路由分组策略 -> 后置策略
func (t *Thoth) Execute(req core.Request) (*core.Response, error) {
	groups := []string{}

	if g := t.beforePolicyGroupName; g != "" {
		groups = append(groups, g)
	}

	if m := t.routeMap.Load(); m != nil {
		routeMap := *m
		// 从路由表中获取目标分组名
		groupNames := routeMap[req.GetEvent()]
		groups = append(groups, groupNames...)
	}

	rsp := &core.Response{}
	errs := []error{}
	for _, g := range groups {
		state, err := t.executor.Exec(g, req, rsp)
		if err != nil {
			errs = append(errs, err)
			// 执行错误 记录错误信息, 继续执行后续分组
		}

		if state.IsHalt {
			// 策略内Halt, 终止策略分组执行, 但还需要执行后置策略
			state.IsHalt = false
			break
		}
	}

	// 只要后置策略配置了, 就无条件执行
	if g := t.afterPolicyGroupName; g != "" {
		_, err := t.executor.Exec(g, req, rsp)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return rsp, errors.Join(errs...)
}
