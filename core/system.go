package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/expr-thoth/expr-thoth/core/ds"
)

type System struct {
	Map   ds.Map
	Slice ds.Slice

	ctx    context.Context
	state  *ThothState
	env    map[string]any
}

func NewSystem(ctx context.Context, state *ThothState, env map[string]any) *System {
	return &System{
		ctx:    ctx,
		state:  state,
		env:    env,
	}
}

// 调试日志函数, 输出日志到控制台
func (s *System) Log(args ...any) bool {
	fmt.Println(args...)
	return true
}

// 当前策略文件结束执行, 继续执行分组内的下一个策略文件
func (s *System) Return() bool {
	s.state.IsReturn = true
	return true
}

// 当前策略分组结束执行, 继续执行下一个分组的策略文件
func (s *System) ExitGroup() bool {
	s.state.IsExitGroup = true
	return true
}

// 当前策略执行中止, 直接返回结果
func (s *System) Halt() bool {
	s.state.IsHalt = true
	return true
}

func (s *System) SetVar(name string, value any) bool {
	if strings.HasPrefix(name, "Var") {
		// 目前仅支持设置以Var开头的变量, 以区分系统内置变量和用户自定义变量
		s.env[name] = value
		return true
	}

	return false

}

func (s *System) Import(varName, plgName string, args ...any) bool {
	plg := s.Load(plgName, args...)
	if plg == nil {
		return false
	}

	s.SetVar(varName, plg)
	return true
}

func (s *System) Load(plgName string, args ...any) any {
	if factory, exists := gPluginFactoryMap[plgName]; exists {
		return factory(s.ctx, args...)
	}
	return nil
}

func (s *System) Sync(f func()) bool {

	return true
}
