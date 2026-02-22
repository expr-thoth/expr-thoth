package core

import (
	"strings"

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
