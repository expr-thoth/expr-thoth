package statement

import (
	"fmt"
	"sync"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// Compiler 负责将表达式字符串编译成可执行的程序, 并提供缓存机制以提高性能
type Compiler struct {
	cache sync.Map
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
func (c *Compiler) GetCached(statement string) (*vm.Program, bool) {
	if cached, ok := c.cache.Load(statement); ok {
		return cached.(*vm.Program), ok
	}
	return nil, false
}

// GetOrCompile 从缓存中获取编译后的程序, 如果缓存中不存在则编译并存储到缓存中
func (c *Compiler) GetOrCompile(statement string, env map[string]any) (*vm.Program, error) {
	if cached, ok := c.GetCached(statement); ok {
		return cached, nil
	}
	program, err := c.Compile(statement, env)
	if err != nil {
		return nil, err
	}
	c.cache.Store(statement, program)
	return program, nil
}

// RunAsBool 运行表达式并将结果转换为bool类型返回
func (c *Compiler) RunAsBool(statement string, env map[string]any) (bool, error) {
	rst, err := c.Run(statement, env)
	if err != nil {
		return false, err
	}

	fVal, ok := rst.(bool)
	if !ok {
		return false, fmt.Errorf("%s is not bool type", statement)
	}

	return fVal, nil
}

// RunWithoutRst 运行表达式并抛弃执行结果
func (c *Compiler) RunWithoutRst(statement string, env map[string]any) error {
	_, err := c.Run(statement, env)
	if err != nil {
		return err
	}

	return nil
}

func (c *Compiler) Run(statement string, env map[string]any) (any, error) {
	program, err := c.GetOrCompile(statement, env)
	if err != nil {
		return nil, err
	}

	rst, err := expr.Run(program, env)
	if err != nil {
		return false, err
	}

	return rst, nil
}
