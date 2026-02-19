package file

import (
	"context"
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-thoth/expr-thoth/core"
	"github.com/expr-thoth/expr-thoth/statement"
)

type Executor struct {
	c     *statement.Compiler
	state *core.ThothState
	env   map[string]any
}

func NewExecutor(ctx context.Context, cmp *statement.Compiler) *Executor {
	state := &core.ThothState{}
	env := map[string]any{}
	sys := core.NewSystem(ctx, state, env)
	env["Sys"] = sys

	return &Executor{
		c:     cmp,
		env:   env,
		state: state,
	}
}

// Exec 执行单个策略文件
// 如果文件中存在错误, 则立刻返回, 不继续执行后续的内容
func (e *Executor) Exec(file *core.PolicyFile, req core.Request, rsp *core.Response) error {
	// 策略未启用不执行
	if file == nil || file.Meta.Enabled == false {
		return nil
	}

	// 灰度不满足不执行
	if req.GetGrayCount() > file.Meta.GrayCount {
		return nil
	}

	e.env["Req"] = req
	e.env["Rsp"] = rsp

	for _, filter := range file.Filter {
		fVal, err := e.execAsBool(filter.When)
		if err != nil {
			return err
		}

		// 过滤条件不满足不执行后续动作
		if !fVal {
			return nil
		}

		// 满足条件执行ThenList中的代码
		for _, then := range filter.ThenList {
			err := e.execWithoutRst(then)
			if err != nil {
				return err
			}

			// Filter不支持Return等操作, 因此不检查
		}

	}

	for _, action := range file.Action {
		program, err := e.c.GetOrCompile(action.When, e.env)
		if err != nil {
			return err
		}

		rst, err := expr.Run(program, e.env)
		if err != nil {
			return err
		}
		aVal, ok := rst.(bool)
		if !ok {
			return fmt.Errorf("action condition result is not bool type")
		}

		if !aVal {
			// Action条件不满足时, 仅不执行当前Action的ThenList
			continue
		}

		for _, then := range action.ThenList {
			err := e.execWithoutRst(then)
			if err != nil {
				return err
			}

			// 终止状态检测
			if e.state.IsReturn || e.state.IsExitGroup || e.state.IsHalt {
				// 结束当前策略文件的执行, 重置Return状态
				// 由上层的执行器判断是否继续执行分组内的其他策略文件
				e.state.IsReturn = false
				return nil
			}
		}
	}

	return nil
}

// 获取当前执行器的状态, 供外部查询
func (e *Executor) GetState() *core.ThothState {
	return e.state
}

func (e *Executor) execAsBool(statement string) (bool, error) {
	program, err := e.c.GetOrCompile(statement, e.env)
	if err != nil {
		return false, err
	}

	rst, err := expr.Run(program, e.env)
	if err != nil {
		return false, err
	}

	fVal, ok := rst.(bool)
	if !ok {
		return false, fmt.Errorf("%s is not bool type", statement)
	}

	return fVal, nil
}

func (e *Executor) execWithoutRst(statement string) error {
	program, err := e.c.GetOrCompile(statement, e.env)
	if err != nil {
		return err
	}

	_, err = expr.Run(program, e.env)
	if err != nil {
		return err
	}

	return nil
}
