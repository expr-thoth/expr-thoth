package statement

import (
	"sync"

	"github.com/expr-lang/expr/vm"
)

// VmProgramCache 是一个线程安全的缓存结构, 用于存储编译后的表达式程序
type VmProgramCache struct {
	mu   sync.RWMutex
	data map[string]*vm.Program
}

// NewCache 创建一个新的 VmProgramCache 实例
func NewCache() *VmProgramCache {
	return &VmProgramCache{
		data: make(map[string]*vm.Program),
	}
}

// Get 从缓存中获取编译后的程序
func (m *VmProgramCache) Get(key string) (*vm.Program, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

// Set 将编译后的程序存储到缓存中
func (m *VmProgramCache) Set(key string, val *vm.Program) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = val
}

// Delete 从缓存中删除指定的程序
func (m *VmProgramCache) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}
