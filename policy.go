package thoth

import (
	"fmt"
	"strings"

	"github.com/expr-thoth/expr-thoth/core"
	"github.com/expr-thoth/expr-thoth/statement"
	"gopkg.in/yaml.v3"
)

// 策略文件对象, 参考 doc/example/expr.yaml 的结构定义
type PolicyFile struct {
	Meta   PolicyFileMeta      `yaml:"Meta"`
	Filter []*PolicyFileFilter `yaml:"Filter"`
	Action []*PolicyFileAction `yaml:"Action"`
}

// 策略文件Meta信息, 包含一些基础属性, 以及灰度和优先级等控制属性
type PolicyFileMeta struct {
	ID          string `yaml:"ID"`
	Name        string `yaml:"Name"`
	GrayCount   int    `yaml:"GrayCount"`
	Priority    int    `yaml:"Priority"`
	Enabled     bool   `yaml:"Enabled"`
	IsAsync     bool   `yaml:"IsAsync"`
	Owner       string `yaml:"Owner"`
	Description string `yaml:"Description"`

	// 业务扩展字段, 可关联任意与策略文件相关的额外信息
	Tag map[string]any
}

// 策略文件中的Filter和Action结构基本相同, Filter模块主要用于条件判断, 不支持Return的执行流控制. Action模型用于执行具体的逻辑, 支持全部操作
type PolicyFileFilter struct {
	Name      string   `yaml:"Name"`
	When      string   `yaml:"When"`
	WhenAll   []string `yaml:"WhenAll"`
	WhenAny   []string `yaml:"WhenAny"`
	ThenList  []string `yaml:"ThenList"`
	ThenGroup []string `yaml:"ThenGroup"`
}

type PolicyFileAction PolicyFileFilter

// 基于YAML文件加载策略代码到内存
// YAML文件格式即为策略规范, 其存储方式由调用方决定, 可以是本地文件, 也可以是远程配置中心等
func NewPolicyFile(yamlStr string) (*PolicyFile, error) {
	var policyFile PolicyFile
	err := yaml.Unmarshal([]byte(yamlStr), &policyFile)
	if err != nil {
		return nil, err
	}

	// When条件处理, WhenAll和WhenAny只是语法糖, 其最终会被转换成When条件, 以便后续执行引擎统一处理
	// 实际策略中, 三种When条件应该只使用一种
	for _, filter := range policyFile.Filter {
		if filter.When != "" {
			continue
		}

		if len(filter.WhenAll) > 0 {
			filter.When = strings.Join(filter.WhenAll, "&&")
			continue
		}

		if len(filter.WhenAny) > 0 {
			filter.When = strings.Join(filter.WhenAny, "||")
			continue
		}
	}

	for _, action := range policyFile.Action {
		if action.When != "" {
			continue
		}

		if len(action.WhenAll) > 0 {
			action.When = strings.Join(action.WhenAll, "&&")
			continue
		}

		if len(action.WhenAny) > 0 {
			action.When = strings.Join(action.WhenAny, "||")
			continue
		}

	}

	return &policyFile, nil
}

// Exec 执行单个策略文件
// 如果文件中存在错误, 则立刻返回, 不继续执行后续的内容
func (f *PolicyFile) Exec(cmp *statement.Compiler, req core.Request, rsp *core.Response, env map[string]any) (*core.ThothState, error) {
	// 策略未启用不执行
	if f == nil || f.Meta.Enabled == false {
		return core.DefaultThothState, nil
	}

	// 灰度不满足不执行
	if req.GetGrayCount() > f.Meta.GrayCount {
		return core.DefaultThothState, nil
	}

	env["Req"] = req
	env["Rsp"] = rsp

	for _, filter := range f.Filter {
		fVal, err := cmp.RunAsBool(filter.When, env)
		if err != nil {
			return nil, fmt.Errorf("exec filter When with err:%v", err)
		}

		// 过滤条件不满足不执行后续动作
		if !fVal {
			return core.DefaultThothState, nil
		}

		// 满足条件执行ThenList中的代码
		for _, then := range filter.ThenList {
			err := cmp.RunWithoutRst(then, env)
			if err != nil {
				return nil, fmt.Errorf("exec filter ThenList with err:%v", err)
			}

			// Filter不支持Return等操作, 因此不检查
		}

	}

	for _, action := range f.Action {
		aVal, err := cmp.RunAsBool(action.When, env)
		if err != nil {
			return nil, fmt.Errorf("exec action When with err: %v", err)
		}

		if !aVal {
			// Action条件不满足时, 仅不执行当前Action的ThenList
			continue
		}

		for _, then := range action.ThenList {
			err := cmp.RunWithoutRst(then, env)
			if err != nil {
				return nil, fmt.Errorf("exec action ThenList with err: %v", err)
			}

			// 终止状态检测
			state := f.getState(env)
			if state.IsReturn || state.IsExitGroup || state.IsHalt {
				// 结束当前策略文件的执行, 重置Return状态
				// 由上层的执行器判断是否继续执行分组内的其他策略文件
				state.IsReturn = false
				return state, nil
			}
		}
	}

	return f.getState(env), nil
}

// // 获取当前执行器的状态, 供外部查询
func (e *PolicyFile) getState(env map[string]any) *core.ThothState {
	return env["ThothState"].(*core.ThothState)
}
