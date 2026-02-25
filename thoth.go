package thoth

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/expr-thoth/expr-thoth/core"
	"github.com/expr-thoth/expr-thoth/router"
)

// Thoth是整个策略引擎的核心, 负责管理策略分组, 路由表以及策略执行器等核心组件
type Thoth struct {
	options  *Options
	routeMap atomic.Pointer[router.RouteMap]
	executor *Executor
}

// NewThoth创建Thoth实例
func NewThoth(ctx context.Context, options ...Option) *Thoth {
	opts := loadOptions(options...)
	return &Thoth{
		executor: NewExecutor(ctx, opts),
		options:  opts,
	}
}

// 注册插件, 可使用Sys变量加载注册的插件
// 初始化阶段调用, 非线程安全
func (t *Thoth) RegisterPlugin(name string, f core.PluginFactory) error {
	core.RegisterPlugin(name, f)
	return nil
}

// 执行Init分组策略, 业务决定具体调用时机, 例如启动时执行一次, 或者Init策略发生变更时执行一次
// 业务需要保证Init分组逻辑幂等
func (t *Thoth) ExecuteInitPolicy() error {
	g := t.options.initPolicyGroupName
	if g == "" {
		return errors.New("init group is empty")
	}

	rsp := &core.Response{}
	_, err := t.executor.Exec(g, core.InitRequest{}, rsp)
	if err != nil {
		return err
	}

	initError := rsp.GetMsg("thoth_init_error")
	if initError != "" {
		return fmt.Errorf("init group err: %v", initError)
	}

	return nil
}

// 注册策略路由表, 执行前置策略后, 按照路由表决定事件后续执行的分组
// 线程安全函数, 支持热加载路由表
func (t *Thoth) RegisterRouteMap(configs []router.RouteConfig) {
	routeMap := router.New(configs)
	t.routeMap.Store(&routeMap)
}

// 设置分组数据, 其存储方式由调用方决定, 可以是本地文件, 也可以是远程配置中心等
// 线程安全函数, 支持热加载分组数据
func (t *Thoth) SetGroup(groupName string, policyFiles []*PolicyFile) {
	t.executor.UpdateGroup(groupName, policyFiles)
}

// 策略入口函数, 其执行流程为: 前置策略 -> 路由分组策略 -> 后置策略
func (t *Thoth) Execute(req core.Request) (*core.Response, error) {
	groups := []string{}

	if g := t.options.beforePolicyGroupName; g != "" {
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
	if g := t.options.afterPolicyGroupName; g != "" {
		_, err := t.executor.Exec(g, req, rsp)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return rsp, errors.Join(errs...)
}
