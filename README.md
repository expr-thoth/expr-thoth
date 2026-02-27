# expr-thoth


## 项目简介

Thoth（托特）是古埃及神话中的智慧之神、知识与书写之神，同时也是魔法、科学和艺术的守护神。expr-thoth 是一个基于表达式的策略引擎，用于灵活管理和执行各种业务策略。它支持通过 YAML 配置文件定义策略，并提供丰富的流程控制、异步执行、灰度发布等功能。

## 核心功能

- **灵活的策略定义**：使用 YAML 文件定义策略，支持条件判断、循环执行、并发执行等
- **流程控制**：支持 Halt、ExitGroup、Return 等流程控制指令
- **异步执行**：支持策略的异步执行，提高系统性能
- **灰度发布**：支持基于灰度值的策略执行控制
- **插件系统**：支持自定义插件扩展功能
- **路由系统**：基于事件类型的策略路由
- **热加载**：支持路由表和策略分组的热加载

## 架构设计

### 核心组件

- **Thoth**：策略引擎的核心，负责管理策略分组、路由表和执行器
- **PolicyFile**：策略文件对象，包含元数据、过滤器和操作
- **Executor**：执行器，负责执行策略
- **Router**：路由表，根据事件类型将请求路由到不同的策略分组

### 执行流程

1. **前置策略**：执行全局前置策略
2. **路由分组策略**：根据事件类型从路由表中获取目标分组，执行对应策略
3. **后置策略**：执行全局后置策略

## 快速开始

### 安装

```bash
go get github.com/expr-thoth/expr-thoth
```

### 基本使用

```go
import (
	"context"
	thoth "github.com/expr-thoth/expr-thoth"
	"github.com/expr-thoth/expr-thoth/router"
)

// 创建 Thoth 实例
thothIns := thoth.NewThoth()

// 注册路由配置
routeConfig := router.RouteConfig{
	EventList: []string{"event_group_test"},
	GroupList: []string{"testGroup"},
	Enabled:   true,
	Priority:  100,
}
thothIns.RegisterRouteMap([]router.RouteConfig{routeConfig})

// 加载策略文件
policyFile, err := thoth.NewPolicyFile(yamlStr)
if err != nil {
	panic(err)
}

// 设置策略分组
thothIns.SetGroup("testGroup", []*thoth.PolicyFile{policyFile})

// 执行策略
ctx := context.Background()
req := &testutil.Request{
	Event:  "event_group_test",
	UserId: "alan turing",
	Attrs:  map[string]string{"role": "admin"},
}

rsp, err := thothIns.Execute(ctx, req)
if err != nil {
	panic(err)
}

// 处理执行结果
fmt.Println(rsp.GetMsg("policy_result"))
```

## 策略文件格式

策略文件使用 YAML 格式定义，主要包含以下部分：

### 基本结构

```yaml
Meta:
  ID: 1001
  Name: 测试策略
  Enabled: true
  GrayCount: 10000
  Priority: 100
  IsAsync: false
  Owner: lizec
  Description: 测试策略描述
Filter:
  - Name: 判断事件
    When: Req.GetEvent() == 'event_test'
    ThenList:
      - Sys.SetVar('VarRole', Req.Attrs['role'])
Action:
  - Name: 管理员权限检查
    When: VarRole == 'admin'
    ThenList:
      - Rsp.SetMsg('policy_result', 'welcome admin')
  - Name: 普通用户处理
    When: true
    ThenList:
      - Rsp.SetMsg('policy_result', 'welcome user')
```

### 支持的表达式

- **Req**：请求对象，包含事件、用户ID、属性等
- **Rsp**：响应对象，用于设置返回消息
- **Sys**：系统对象，提供流程控制、变量设置等功能
- **自定义变量**：通过 Sys.SetVar 设置的变量

### 流程控制指令

- **Sys.Halt()**：终止所有策略执行
- **Sys.ExitGroup()**：终止当前分组的策略执行
- **Sys.Return()**：终止当前策略文件的执行

### 并发执行

使用 `ThenGroup` 可以并发执行多个表达式：

```yaml
Action:
  - Name: 并发执行
    When: true
    ThenGroup:
      - Rsp.SetMsg('A', '1')
      - Rsp.SetMsg('B', '2')
      - Rsp.SetMsg('C', '3')
```

### 循环执行

使用 `ForEach` 可以遍历集合执行表达式：

```yaml
Action:
  - Name: 循环执行
    When: true
    ForEach:
      Item: [1, 2, 3, 4, 5]
      When: v > 2
      ThenList:
        - Rsp.SetMsg('item_'+k, v)
```

## 示例

### 基础示例

查看 `doc/example/expr.yaml` 了解基础策略定义。

### 流程控制示例

查看 `doc/example/flow/` 目录下的示例，了解流程控制的使用。

### 灰度发布示例

查看 `doc/example/gray/` 目录下的示例，了解灰度发布的使用。

### 并发执行示例

查看 `doc/example/then_group.yaml` 了解并发执行的使用。

## API 文档

### Thoth 核心方法

- **NewThoth(options ...Option) *Thoth**：创建 Thoth 实例
- **RegisterPlugin(name string, f core.PluginFactory) error**：注册插件
- **RegisterRouteMap(configs []router.RouteConfig)**：注册路由配置
- **SetGroup(groupName string, policyFiles []*PolicyFile)**：设置策略分组
- **Execute(ctx context.Context, req core.Request) (*core.Response, error)**：执行策略
- **ExecuteInitPolicy(ctx context.Context) error**：执行初始化策略

### 配置选项

- **WithBeforePolicy(groupName string)**：设置前置策略分组
- **WithAfterPolicy(groupName string)**：设置后置策略分组
- **WithInitPolicy(groupName string)**：设置初始化策略分组

