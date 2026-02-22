package group

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/expr-thoth/expr-thoth/core"
	"github.com/expr-thoth/expr-thoth/file"
	"github.com/expr-thoth/expr-thoth/statement"
)

type Executor struct {
	ctx context.Context
	// 全局共享编译器, 内部包含表达式缓存
	cmp    *statement.Compiler
	groups sync.Map
}

func NewExecutor(ctx context.Context) *Executor {
	return &Executor{
		ctx: ctx,
		cmp: statement.NewCompiler(ctx),
	}
}

func (e *Executor) UpdateGroup(groupName string, policyFiles []*core.PolicyFile) {
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

	policyGroup, ok := raw.(*PolicyGroup)
	if !ok {
		// 类型错误属于编程错误, 直接panic暴露问题
		panic(fmt.Sprintf("group %s is not PolicyGroup type", groupName))
	}

	state := core.DefaultThothState
	for _, policyFile := range policyGroup.policyFiles {
		// 每个文件使用单独的执行器, 以保证状态隔离
		exector := file.NewExecutor(e.ctx, e.cmp)

		err := exector.Exec(policyFile, req, rsp)
		if err != nil {
			errs = append(errs, fmt.Errorf("group %v with err: %v", policyGroup.Name, err))
			// 执行错误继续执行后续策略, 以保证单个策略错误不影响整体策略执行
		}

		state = exector.GetState()
		if state.IsExitGroup {
			// 结束整个分组的执行, 重置ExitGroup状态, 直接返回当前状态结果
			state.IsExitGroup = false
			return state, errors.Join(errs...)
		}
	}

	// 注意: 由于产生错误依然继续执行, 因此返回前需要收集已经产生的错误并一同返回
	return state, errors.Join(errs...)
}
