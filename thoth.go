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
func NewThoth(options ...Option) *Thoth {
	opts := loadOptions(options...)
	return &Thoth{
		executor: NewExecutor(opts),
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
func (t *Thoth) ExecuteInitPolicy(ctx context.Context) error {
	g := t.options.initPolicyGroupName
	if g == "" {
		return errors.New("init group is empty")
	}

	rsp := &core.Response{}
	// Init策略不支持流程控制指令, 不支持异步策略. 无需处理state和异步策略文件
	_, _, err := t.executor.Exec(ctx, g, core.InitRequest{}, rsp)
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
func (t *Thoth) Execute(ctx context.Context, req core.Request) (*core.Response, error) {
	rsp := &core.Response{}
	errs := []error{}
	var asyncFiles []*PolicyFile
	var isHalt bool

	// 执行前置策略
	if g := t.options.beforePolicyGroupName; g != "" {
		state, files, err := t.executor.Exec(ctx, g, req, rsp)
		if err != nil {
			errs = append(errs, err)
		}
		if len(files) > 0 {
			asyncFiles = files
		}
		if state.IsHalt {
			isHalt = true
		}
	}

	// 执行路由分组策略
	if !isHalt {
		if m := t.routeMap.Load(); m != nil {
			routeMap := *m
			// 从路由表中获取目标分组名
			groupNames := routeMap[req.GetEvent()]
			for _, g := range groupNames {
				state, files, err := t.executor.Exec(ctx, g, req, rsp)
				if err != nil {
					errs = append(errs, err)
				}
				if len(files) > 0 {
					if asyncFiles == nil {
						asyncFiles = files
					} else {
						asyncFiles = append(asyncFiles, files...)
					}
				}
				if state.IsHalt {
					isHalt = true
					break
				}
			}
		}
	}

	// 只要后置策略配置了, 就无条件执行
	if g := t.options.afterPolicyGroupName; g != "" {
		_, files, err := t.executor.Exec(ctx, g, req, rsp)
		if err != nil {
			errs = append(errs, err)
		}
		if len(files) > 0 {
			if asyncFiles == nil {
				asyncFiles = files
			} else {
				asyncFiles = append(asyncFiles, files...)
			}
		}
	}

	if len(asyncFiles) > 0 {
		// 策略异步执行（解耦主链路耗时+保障并发安全）
		// 1. 主链路逻辑完成后异步执行策略, 不阻塞主链路返回, 降低接口耗时
		// 2. asyncFiles已按策略优先级排序, 遍历执行可保证策略执行顺序符合预期
		// 3. 统一在单个异步协程中执行所有策略, 避免多协程并发读写req/env等变量
		// 4. 使用WithoutCancel创建新ctx, 脱离主链路超时控制, 确保异步策略完整执行
		// 5. 异步策略执行时, 主链路已结束, 因此不可访问rsp
		newCtx := context.WithoutCancel(ctx)
		asyncPolicyPool.Submit(func() {
			for _, policyFile := range asyncFiles {
				env := MakeEnv(newCtx)
				_, err := policyFile.Exec(t.options.cmp, req, nil, env)
				if err != nil {
					t.options.logger.ErrorMetrics(newCtx, "Thoth.Executor.Exec.Error", "async policy %s exec with err: %v", policyFile.Meta.ID, err)
				}
			}
		})
	}

	return rsp, errors.Join(errs...)
}
