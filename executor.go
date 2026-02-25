package thoth

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/expr-thoth/expr-thoth/core"
	"github.com/panjf2000/ants/v2"
)

var asyncPolicyPool, _ = ants.NewPool(2000, ants.WithNonblocking(false))

type Executor struct {
	ctx     context.Context
	groups  sync.Map
	options *Options
}

type PolicyGroup struct {
	Name        string
	policyFiles []*PolicyFile
}

func NewExecutor(ctx context.Context, opts *Options) *Executor {
	return &Executor{
		ctx:     ctx,
		options: opts,
	}
}

func (e *Executor) UpdateGroup(groupName string, policyFiles []*PolicyFile) {
	// 按照优先级从高到低排序, 优先级数值越大优先级越高
	sort.Slice(policyFiles, func(i, j int) bool {
		return policyFiles[i].Meta.Priority > policyFiles[j].Meta.Priority
	})

	e.groups.Store(groupName, &PolicyGroup{
		Name:        groupName,
		policyFiles: policyFiles,
	})
}

// Exec 执行指定分组的策略, 返回分组执行结果和执行过程中遇到的错误列表
// 如果分组内某个策略文件执行错误, 记录错误信息, 继续执行后续策略文件, 最终返回所有错误信息
func (e *Executor) Exec(groupName string, req core.Request, rsp *core.Response) (*core.ThothState, error) {
	var errs []error
	raw, ok := e.groups.Load(groupName)
	if !ok || raw == nil {
		return nil, fmt.Errorf("group %s not found", groupName)
	}

	policyGroup := raw.(*PolicyGroup)

	state := core.DefaultThothState
	asyncFile := make([]*PolicyFile, 0, len(policyGroup.policyFiles))
	for _, policyFile := range policyGroup.policyFiles {
		if policyFile.Meta.IsAsync {
			// 收集异步策略文件, 最后提交到协程池执行
			asyncFile = append(asyncFile, policyFile)
		}

		// 每个文件使用单独的执行器, 以保证状态隔离
		env := MakeEnv(e.ctx)
		state, err := policyFile.Exec(e.options.cmp, req, rsp, env)

		if err != nil {
			errs = append(errs, fmt.Errorf("exec group %v with err: %v", policyGroup.Name, err))
			// 执行错误继续执行后续策略, 以保证单个策略错误不影响整体策略执行
		}

		if state.IsExitGroup || state.IsHalt {
			// 结束整个分组的执行, 重置ExitGroup状态, 直接返回当前状态结果
			state.IsExitGroup = false
			return state, errors.Join(errs...)
		}
	}

	// TODO: 分组级别也不行, 要提升到最外层, 整个主链路执行完毕了才能结束
	// 最后在新协程中执行所有异步策略, 只需要保证主线程正常返回即可, 避免直接并行执行产生并发问题
	// 需要复制ctx, 取消原有主链路的超时控制逻辑
	newCtx := context.WithoutCancel(e.ctx)
	asyncPolicyPool.Submit(func() {
		for _, policyFile := range asyncFile {
			// recover可能得panic, 避免异步协程中的panic导致整个服务终止
			defer recoverPanic(newCtx, e.options.logger)

			env := MakeEnv(newCtx)
			_, err := policyFile.Exec(e.options.cmp, req, rsp, env)
			if err != nil {
				e.options.logger.ErrorMetrics(newCtx, "Thoth.Executor.Exec.Error", "async policy %s exec with err: %v", policyFile.Meta.ID, err)
			}
		}
	})

	// 注意: 由于产生错误依然继续执行, 因此返回前需要收集已经产生的错误并一同返回
	return state, errors.Join(errs...)
}
