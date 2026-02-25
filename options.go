package thoth

import (
	"github.com/expr-thoth/expr-thoth/core"
	"github.com/expr-thoth/expr-thoth/statement"
)

var (
	defaultLogger = core.SimpleThothLogger{}
)

type Option func(opts *Options)

func loadOptions(options ...Option) *Options {
	opts := new(Options)
	for _, option := range options {
		option(opts)
	}

	if opts.cmp == nil {
		opts.cmp = &statement.Compiler{}
	}

	if opts.logger == nil {
		opts.logger = defaultLogger
	}

	return opts
}

type Options struct {
	initPolicyGroupName   string
	beforePolicyGroupName string
	afterPolicyGroupName  string
	logger                core.ThothLogger

	// 全局共享编译器, 内部包含表达式缓存
	cmp *statement.Compiler
}

func WithOptions(options Options) Option {
	return func(opts *Options) {
		*opts = options
	}
}

// 设置Init分组策略, 可在该分组配置仅需单次执行, 与请求无关的初始化相关的逻辑
func WithInitPolicy(groupName string) Option {
	return func(opts *Options) {
		opts.initPolicyGroupName = groupName
	}
}

// 设置前置策略分组名, 执行策略时所有事件始终先执行前置策略
func WithBeforePolicy(groupName string) Option {
	return func(opts *Options) {
		opts.beforePolicyGroupName = groupName
	}
}

// 设置后置策略分组名, 其余策略执行完毕后, 最后执行后置策略
// 无论策略中是否发生错误或者执行控制流指令, 只要后置策略配置了, 就无条件执行
func WithAfterPolicy(groupName string) Option {
	return func(opts *Options) {
		opts.afterPolicyGroupName = groupName
	}
}

// 设置日志组件
func WithLogger(logger core.ThothLogger) Option {
	return func(opts *Options) {
		opts.logger = logger
	}
}
