package statement

import (
	"context"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Compiler 负责将表达式字符串编译成可执行的程序, 并提供缓存机制以提高性能
type Compiler struct {
	ctx   context.Context
	cache VmProgramCache
}

// NewCompiler 创建一个新的 Compiler 实例, 需要传入上下文对象以便在编译过程中进行日志记录等操作
func NewCompiler(ctx context.Context) *Compiler {
	return &Compiler{
		ctx: ctx,
		cache: VmProgramCache{
			data: make(map[string]*vm.Program),
		},
	}
}

// Compile 将表达式字符串编译成可执行的程序, 需要传入环境变量以便在编译过程中进行变量解析
func (c *Compiler) Compile(statement string, env map[string]any) (*vm.Program, error) {
	program, err := expr.Compile(statement, expr.Env(env), expr.AllowUndefinedVariables())
	if err != nil {
		return nil, err
	}
	return program, nil
}

// GetCached 从缓存中获取编译后的程序, 如果缓存中不存在则返回 nil
func (c *Compiler) GetCached(statement string, env map[string]any) (*vm.Program, error) {
	if cached, ok := c.cache.Get(statement); ok {
		return cached, nil
	}
	return nil, nil
}

// GetOrCompile 从缓存中获取编译后的程序, 如果缓存中不存在则编译并存储到缓存中
func (c *Compiler) GetOrCompile(statement string, env map[string]any) (*vm.Program, error) {
	if cached, ok := c.cache.Get(statement); ok {
		return cached, nil
	}
	program, err := c.Compile(statement, env)
	if err != nil {
		return nil, err
	}
	c.cache.Set(statement, program)
	return program, nil
}
