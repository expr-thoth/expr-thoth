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
	groups  sync.Map
	options *Options
}

type PolicyGroup struct {
	Name        string
	policyFiles []*PolicyFile
}

func NewExecutor(opts *Options) *Executor {
	return &Executor{
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
func (e *Executor) Exec(ctx context.Context, groupName string, req core.Request, rsp *core.Response) (*core.ThothState, []*PolicyFile, error) {
	var errs []error
	raw, ok := e.groups.Load(groupName)
	if !ok || raw == nil {
		return nil, nil, fmt.Errorf("group %s not found", groupName)
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
		env := MakeEnv(ctx)
		state, err := policyFile.Exec(e.options.cmp, req, rsp, env)

		if err != nil {
			errs = append(errs, fmt.Errorf("exec group %v with err: %v", policyGroup.Name, err))
			// 执行错误继续执行后续策略, 以保证单个策略错误不影响整体策略执行
			continue
		}

		if state == nil {
			// 状态变量无效, 也跳过
			continue
		}

		if state.IsExitGroup || state.IsHalt {
			// 结束整个分组的执行, 重置ExitGroup状态, 直接返回当前状态结果
			state.IsExitGroup = false
			return state, asyncFile, errors.Join(errs...)
		}
	}

	// 注意: 由于产生错误依然继续执行, 因此返回前需要收集已经产生的错误并一同返回
	return state, asyncFile, errors.Join(errs...)
}
