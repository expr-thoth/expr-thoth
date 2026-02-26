package thoth

import (
	"fmt"
	"reflect"
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
	ForEach   *ForEach `yaml:"ForEach"`
}

type PolicyFileAction PolicyFileFilter

type ForEach struct {
	Item     string   `yaml:"Item"`
	When     string   `yaml:"When"`
	WhenAll  []string `yaml:"WhenAll"`
	WhenAny  []string `yaml:"WhenAny"`
	ThenList []string `yaml:"ThenList"`
}

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
		filter.When = mergeWhen(filter.When, filter.WhenAll, filter.WhenAny)
	}

	for _, action := range policyFile.Action {
		action.When = mergeWhen(action.When, action.WhenAll, action.WhenAny)
		if action.ForEach != nil {
			action.ForEach.When = mergeWhen(action.ForEach.When, action.ForEach.WhenAll, action.ForEach.WhenAny)
		}
	}

	return &policyFile, nil
}

func mergeWhen(when string, whenAll, whenAny []string) string {
	if when != "" {
		return when
	}

	if len(whenAll) > 0 {
		return strings.Join(whenAll, "&&")
	}

	if len(whenAny) > 0 {
		return strings.Join(whenAny, "||")
	}

	return ""
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

	// 处理 Filter 阶段
	if state, err := f.processFilters(cmp, env); err != nil || state != nil {
		return state, err
	}

	// 处理 Action 阶段
	if state, err := f.processActions(cmp, env); err != nil || state != nil {
		return state, err
	}

	return f.getState(env), nil
}

// processFilters 处理所有 Filter，返回需要提前终止的状态或错误
func (f *PolicyFile) processFilters(cmp *statement.Compiler, env map[string]any) (*core.ThothState, error) {
	for _, filter := range f.Filter {
		// 执行条件判断
		ok, err := execCondition(cmp, filter.When, env)
		if err != nil {
			return nil, fmt.Errorf("exec filter When: %w", err)
		}
		if !ok {
			return core.DefaultThothState, nil // 条件不满足, 策略文件停止执行
		}

		// 执行 ThenList 并检查是否需要提前返回
		if state, err := f.execThenListWithState(cmp, filter.ThenList, env); err != nil || state != nil {
			return state, err
		}
	}
	return nil, nil
}

// processActions 处理所有 Action，返回需要提前终止的状态或错误
func (f *PolicyFile) processActions(cmp *statement.Compiler, env map[string]any) (*core.ThothState, error) {
	for _, action := range f.Action {
		// 判断当前 Action 是否满足条件
		ok, err := execCondition(cmp, action.When, env)
		if err != nil {
			return nil, fmt.Errorf("exec action When: %w", err)
		}
		if !ok {
			// 条件不满足, 继续执行下一个Action
			continue
		}

		// 执行 ThenList
		if state, err := f.execThenListWithState(cmp, action.ThenList, env); err != nil || state != nil {
			return state, err
		}

		// 执行 ThenGroup
		// 并发执行, 不检查状态
		if err := execThenGroup(cmp, action.ThenGroup, env); err != nil {
			return nil, fmt.Errorf("exec action ThenGroup: %w", err)
		}

		// 执行 ForEach
		if action.ForEach != nil {
			if state, err := f.execForEach(cmp, action.ForEach, env); err != nil || state != nil {
				return state, err
			}
		}
	}
	return nil, nil
}

// execCondition 执行条件语句并返回布尔结果
func execCondition(cmp *statement.Compiler, when string, env map[string]any) (bool, error) {
	val, err := cmp.RunAsBool(when, env)
	if err != nil {
		return false, err
	}
	return val, nil
}

// execThenList 执行一个 ThenList，返回是否需要提前终止（bool）和错误
func (f *PolicyFile) execThenList(cmp *statement.Compiler, thenList []string, env map[string]any) (bool, error) {
	for _, then := range thenList {
		if err := cmp.RunWithoutRst(then, env); err != nil {
			return false, err
		}
		state := f.getState(env)
		if state.IsReturn || state.IsExitGroup || state.IsHalt {
			state.IsReturn = false // 重置Return标志
			return true, nil
		}
	}
	return false, nil
}

// execThenListWithState 执行 ThenList，若需提前终止则直接返回状态
func (f *PolicyFile) execThenListWithState(cmp *statement.Compiler, thenList []string, env map[string]any) (*core.ThothState, error) {
	needReturn, err := f.execThenList(cmp, thenList, env)
	if err != nil {
		return nil, err
	}
	if needReturn {
		return f.getState(env), nil
	}
	return nil, nil
}

// execThenGroup 并发执行 ThenGroup
func execThenGroup(cmp *statement.Compiler, thenGroup []string, env map[string]any) error {
	if len(thenGroup) == 0 {
		return nil
	}

	tasks := make([]Task, 0, len(thenGroup))
	for _, then := range thenGroup {
		t := then
		tasks = append(tasks, func() error {
			return cmp.RunWithoutRst(t, env)
		})
	}
	return GoAndWait(tasks...)
}

// execForEach 处理 ForEach 逻辑
func (f *PolicyFile) execForEach(cmp *statement.Compiler, fe *ForEach, env map[string]any) (*core.ThothState, error) {
	itemRaw, err := cmp.Run(fe.Item, env)
	if err != nil {
		return nil, fmt.Errorf("exec ForEach Item: %w", err)
	}

	// 使用反射统一处理不同类型的集合
	rv := reflect.ValueOf(itemRaw)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			k := i
			v := rv.Index(i).Interface()
			if state, err := f.execForEachItem(cmp, fe, k, v, env); err != nil || state != nil {
				return state, err
			}
		}
	case reflect.Map:
		iter := rv.MapRange()
		for iter.Next() {
			k := iter.Key().Interface()
			v := iter.Value().Interface()
			if state, err := f.execForEachItem(cmp, fe, k, v, env); err != nil || state != nil {
				return state, err
			}
		}
	default:
		return nil, fmt.Errorf("ForEach item must be slice, array or map, got %T", itemRaw)
	}
	return nil, nil
}

// execForEachItem 处理 ForEach 中的单个元素
func (f *PolicyFile) execForEachItem(cmp *statement.Compiler, fe *ForEach, k any, v any, env map[string]any) (*core.ThothState, error) {
	env["k"] = k
	env["v"] = v

	ok, err := execCondition(cmp, fe.When, env)
	if err != nil {
		return nil, fmt.Errorf("exec ForEach When: %w", err)
	}
	if !ok {
		return nil, nil
	}

	// 执行该元素对应的 ThenList
	return f.execThenListWithState(cmp, fe.ThenList, env)
}

func (f *PolicyFile) getState(env map[string]any) *core.ThothState {
	return env["ThothState"].(*core.ThothState)
}
