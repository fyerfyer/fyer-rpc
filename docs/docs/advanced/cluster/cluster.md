# Cluster

集群管理是fyerrpc框架中的高级功能，根据不同条件和策略管理和路由服务请求。

## 分组管理 (Group)

分组管理将服务实例划分为不同的逻辑组，便于进行灰度发布、A/B测试和服务隔离等场景。

### 基础概念

#### Group 接口

`Group` 接口定义了服务分组的基本行为：

```go
type Group interface {
    // Name 返回分组名称
    Name() string

    // Match 检查实例是否匹配该分组
    Match(instance *naming.Instance) bool

    // Select 从实例列表中选择匹配该分组的实例
    Select(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)
}
```

#### 分组配置

分组通过 Config 结构进行配置：

```go
type Config struct {
    // Name 分组名称
    Name string `json:"name"`

    // Type 分组类型，如A/B测试、金丝雀发布等
    Type string `json:"type"`

    // Matcher 分组匹配规则
    Matcher *MatchConfig `json:"matcher"`

    // Weight 分组权重(0-100)
    Weight int `json:"weight"`

    // EnableHealthCheck 是否启用健康检查
    EnableHealthCheck bool `json:"enable_health_check"`

    // HealthCheckInterval 健康检查间隔
    HealthCheckInterval time.Duration `json:"health_check_interval"`

    // Metadata 分组元数据
    Metadata map[string]string `json:"metadata"`
}
```

#### 匹配规则

fyerrpc支持多种匹配规则，通过 `MatchConfig` 结构配置：

```go
type MatchConfig struct {
    // MatchType 匹配类型：exact(精确匹配)、prefix(前缀匹配)、regex(正则匹配)
    MatchType string `json:"match_type"`

    // MatchKey 匹配的键，如环境变量名、Header名等
    MatchKey string `json:"match_key"`

    // MatchValue 匹配的值
    MatchValue string `json:"match_value"`

    // Labels 标签匹配规则
    Labels map[string]string `json:"labels"`
}
```

### 创建与使用分组

#### 创建分组

使用 `NewGroup` 函数创建分组：

```go
// 创建基于精确匹配的分组
group, err := group.NewGroup("prod-group",
    group.WithMatcher(&group.MatchConfig{
        MatchType:  "exact",
        MatchKey:   "env",
        MatchValue: "production",
    }),
    group.WithWeight(80),
)
if err != nil {
    log.Fatalf("Failed to create group: %v", err)
}

// 创建基于正则匹配的分组
canaryGroup, err := group.NewGroup("canary-group",
    group.WithMatcher(&group.MatchConfig{
        MatchType:  "regex",
        MatchKey:   "version",
        MatchValue: "^2\\..*$", // 匹配所有2.x版本
    }),
    group.WithWeight(20),
)
```

#### 分组匹配

使用分组对实例进行匹配：

```go
// 创建测试实例
instance := &naming.Instance{
    ID:      "instance-1",
    Service: "my-service",
    Address: "localhost:8080",
    Metadata: map[string]string{
        "env": "production",
    },
}

// 检查实例是否匹配分组
if group.Match(instance) {
    fmt.Println("Instance matches production group")
}
```

#### 分组选择

从实例列表中选择匹配分组的实例：

```go
// 创建多个实例
instances := []*naming.Instance{
    {
        ID:      "instance-1",
        Service: "my-service",
        Address: "localhost:8081",
        Metadata: map[string]string{
            "env": "production",
        },
    },
    {
        ID:      "instance-2",
        Service: "my-service",
        Address: "localhost:8082",
        Metadata: map[string]string{
            "env": "testing",
        },
    },
}

// 选择匹配的实例
selected, err := group.Select(context.Background(), instances)
if err != nil {
    log.Fatalf("Group selection failed: %v", err)
}

fmt.Printf("Selected %d instances\n", len(selected))
for _, inst := range selected {
    fmt.Printf("  - %s (%s)\n", inst.ID, inst.Address)
}
```

### A/B测试分组

fyerrpc提供了专门的A/B测试分组实现：

```go
// 创建A/B测试分组选择器
abSelector := group.NewABGroupSelector(
    "ab-test",       // 选择器名称
    "version-a",     // A组名称
    "version-b",     // B组名称
    0.8,             // A组流量比例(80%)
)

// 使用选择器选择实例
ctx := context.Background()
selected, err := abSelector.Select(ctx, instances)
```

您还可以在上下文中指定强制分组：

```go
// 强制使用A组
ctx := context.WithValue(context.Background(), "ab_group", "version-a")
selected, err := abSelector.Select(ctx, instances)
```

## 路由规则 (Router)

路由规则允许您根据请求上下文信息动态选择服务实例，实现请求级别的流量控制。

### 路由接口

路由通过 `GroupRouter` 接口实现：

```go
type GroupRouter interface {
    // Route 根据上下文路由到指定分组
    Route(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)
}
```

### 使用内置路由器

fyerrpc提供了多种内置路由器实现：

#### 标签路由器

基于元数据标签的路由：

```go
// 创建基于标签的路由器
tagRouter := group.NewTagRouter("env", "production")

// 路由实例
instances := getServiceInstances() // 获取服务实例列表
routed, err := tagRouter.Route(context.Background(), instances)
if err != nil {
    log.Fatalf("Routing failed: %v", err)
}
```

#### 权重路由器

基于权重的路由：

```go
// 创建基于权重的路由器，只选择权重>=80的实例
weightRouter := group.NewWeightRouter("weight", 80)

// 路由实例
routed, err := weightRouter.Route(context.Background(), instances)
```

#### 版本路由器

基于服务版本的路由：

```go
// 创建基于版本的路由器
versionRouter := group.NewVersionRouter("1.0.0")

// 路由实例
routed, err := versionRouter.Route(context.Background(), instances)
```

### 路由链

fyerrpc支持路由链，将多个路由规则组合使用：

```go
// 创建路由链
chain := group.NewRouterChain(
    group.NewTagRouter("env", "production"),     // 先按环境路由
    group.NewWeightRouter("weight", 80),         // 再按权重路由
    group.NewVersionRouter("1.0.0"),             // 最后按版本路由
)

// 执行路由链
routed, err := chain.Route(context.Background(), instances)
```

路由链会依次执行每个路由规则，前一个路由的结果将作为下一个路由的输入。如果任何一个路由规则返回空结果或错误，整个路由链将失败。

### 基于上下文的路由

可以通过上下文传递路由信息：

```go
// 设置上下文中的分组信息
ctx := group.WithGroup(context.Background(), group.GroupKey("production"))

// 创建路由器
router := group.NewRouter(groupManager)

// 根据上下文路由
routed, err := router.Route(ctx, instances)
```

## 选择器 (Selector)

选择器负责从一组服务实例中根据特定策略选择合适的实例子集，常用于服务发现和负载均衡场景。

### 选择器接口

选择器通过 `Selector` 接口实现：

```go
type Selector interface {
    // Select 从服务实例列表中选择符合条件的实例
    Select(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error)
    
    // Name 返回选择器名称
    Name() string
}
```

### 选择器配置

选择器通过 Config 结构进行配置：

```go
type Config struct {
    // Strategy 选择策略，如group、version、region等
    Strategy string `json:"strategy"`

    // Filter 筛选规则
    Filter map[string]string `json:"filter"`

    // Priority 选择器优先级，数字越小优先级越高
    Priority int `json:"priority"`

    // Required 是否必须满足选择条件
    Required bool `json:"required"`

    // Metadata 额外的元数据
    Metadata map[string]string `json:"metadata"`
}
```

### 使用内置选择器

fyerrpc提供了多种内置选择器实现：

#### 分组选择器

基于分组信息选择实例：

```go
// 创建分组选择器
groupSelector := selector.NewGroupSelector(
    "group-selector", // 选择器名称
    "group",          // 分组元数据键
    "production",     // 默认分组名
)

// 使用选择器
selected, err := groupSelector.Select(context.Background(), instances)
if err != nil {
    log.Fatalf("Selection failed: %v", err)
}
```

可以通过上下文指定分组：

```go
// 指定使用testing分组
ctx := selector.WithGroup(context.Background(), "group", "testing")
selected, err := groupSelector.Select(ctx, instances)
```

#### 基础选择器

使用自定义选择函数创建选择器：

```go
// 创建基础选择器，选择前三个实例
baseSelector := selector.NewBaseSelector("top3", func(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
    if len(instances) <= 3 {
        return instances, nil
    }
    return instances[:3], nil
})

// 使用选择器
selected, err := baseSelector.Select(context.Background(), instances)
```

### 选择器链

选择器链允许组合多个选择器，更精细地控制实例选择过程：

```go
// 创建选择器链
chain := selector.NewChain("my-chain",
    groupSelector,                // 先按分组选择
    regionSelector,               // 再按区域选择
    healthSelector,               // 最后只保留健康实例
)

// 使用选择器链
selected, err := chain.Select(context.Background(), instances)
```

也可以使用构建器模式创建选择器链：

```go
// 使用构建器创建选择器链
chain := selector.NewChainBuilder().
    Add(groupSelector).
    Add(regionSelector).
    Add(healthSelector).
    Build()

// 使用选择器链
selected, err := chain.Select(context.Background(), instances)
```

## 集成示例

以下是一个综合示例，展示如何结合分组、路由和选择器实现灰度发布：

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/fyerfyer/fyer-rpc/cluster/group"
    "github.com/fyerfyer/fyer-rpc/cluster/selector"
    "github.com/fyerfyer/fyer-rpc/naming"
)

func main() {
    // 创建服务实例
    instances := []*naming.Instance{
        {
            ID:      "instance-1",
            Service: "user-service",
            Version: "1.0.0",
            Address: "10.0.1.1:8080",
            Metadata: map[string]string{
                "env":    "production",
                "region": "us-east",
                "weight": "100",
                "group":  "stable",
            },
        },
        {
            ID:      "instance-2",
            Service: "user-service",
            Version: "1.1.0",
            Address: "10.0.1.2:8080",
            Metadata: map[string]string{
                "env":    "production",
                "region": "us-east",
                "weight": "80",
                "group":  "canary",
            },
        },
        {
            ID:      "instance-3",
            Service: "user-service",
            Version: "1.0.0",
            Address: "10.0.2.1:8080",
            Metadata: map[string]string{
                "env":    "production",
                "region": "us-west",
                "weight": "100",
                "group":  "stable",
            },
        },
    }

    // 1. 创建分组
    stableGroup, err := group.NewGroup("stable",
        group.WithMatcher(&group.MatchConfig{
            MatchType:  "exact",
            MatchKey:   "group",
            MatchValue: "stable",
        }),
        group.WithWeight(90),
    )
    if err != nil {
        log.Fatalf("Failed to create stable group: %v", err)
    }

    canaryGroup, err := group.NewGroup("canary",
        group.WithMatcher(&group.MatchConfig{
            MatchType:  "exact",
            MatchKey:   "group",
            MatchValue: "canary",
        }),
        group.WithWeight(10),
    )
    if err != nil {
        log.Fatalf("Failed to create canary group: %v", err)
    }

    // 2. 创建分组管理器 (这里使用一个简单的实现)
    groupManager := &SimpleGroupManager{
        groups: map[string]group.Group{
            "stable": stableGroup,
            "canary": canaryGroup,
        },
    }

    // 3. 创建路由器
    router := group.NewRouter(groupManager)

    // 4. 创建区域选择器
    regionSelector := selector.NewBaseSelector("region-selector", func(ctx context.Context, insts []*naming.Instance) ([]*naming.Instance, error) {
        region, ok := ctx.Value("region").(string)
        if !ok {
            return insts, nil
        }

        var selected []*naming.Instance
        for _, inst := range insts {
            if r, ok := inst.Metadata["region"]; ok && r == region {
                selected = append(selected, inst)
            }
        }
        
        if len(selected) == 0 {
            return insts, nil // 如果没有匹配的实例，返回所有实例
        }
        return selected, nil
    })

    // 5. 创建选择器链
    selectorChain := selector.NewChain("region-group-chain", regionSelector)

    // 6. 用户请求场景模拟
    
    // 场景1: 稳定版用户，东区
    fmt.Println("Scenario 1: Stable user in East region")
    ctx1 := context.Background()
    ctx1 = group.WithGroup(ctx1, group.GroupKey("stable"))
    ctx1 = context.WithValue(ctx1, "region", "us-east")
    
    // 先路由到分组
    groupInstances, err := router.Route(ctx1, instances)
    if err != nil {
        log.Printf("Routing error: %v", err)
    } else {
        // 再按区域选择
        selected, err := selectorChain.Select(ctx1, groupInstances)
        if err != nil {
            log.Printf("Selection error: %v", err)
        } else {
            fmt.Printf("Selected %d instances:\n", len(selected))
            for _, inst := range selected {
                fmt.Printf("  - %s (%s) [%s, %s]\n", 
                    inst.ID, inst.Address, 
                    inst.Metadata["group"], inst.Metadata["region"])
            }
        }
    }
    
    // 场景2: 灰度用户
    fmt.Println("\nScenario 2: Canary user")
    ctx2 := group.WithGroup(context.Background(), group.GroupKey("canary"))
    
    groupInstances, err = router.Route(ctx2, instances)
    if err != nil {
        log.Printf("Routing error: %v", err)
    } else {
        fmt.Printf("Selected %d instances:\n", len(groupInstances))
        for _, inst := range groupInstances {
            fmt.Printf("  - %s (%s) [%s]\n", 
                inst.ID, inst.Address, inst.Metadata["group"])
        }
    }
}

// 简单分组管理器实现
type SimpleGroupManager struct {
    groups map[string]group.Group
}

func (m *SimpleGroupManager) RegisterGroup(g group.Group) error {
    m.groups[g.Name()] = g
    return nil
}

func (m *SimpleGroupManager) GetGroup(name string) (group.Group, error) {
    if g, ok := m.groups[name]; ok {
        return g, nil
    }
    return nil, fmt.Errorf("group not found: %s", name)
}

func (m *SimpleGroupManager) ListGroups() []group.Group {
    groups := make([]group.Group, 0, len(m.groups))
    for _, g := range m.groups {
        groups = append(groups, g)
    }
    return groups
}
```

## 高级使用场景

### 灰度发布

使用分组和权重实现灰度发布：

```go
// 创建稳定版和灰度版分组
stableGroup, _ := group.NewGroup("stable", 
    group.WithMatcher(&group.MatchConfig{
        MatchType:  "exact",
        MatchKey:   "version",
        MatchValue: "1.0.0",
    }),
    group.WithWeight(90),
)

canaryGroup, _ := group.NewGroup("canary", 
    group.WithMatcher(&group.MatchConfig{
        MatchType:  "exact",
        MatchKey:   "version",
        MatchValue: "1.1.0",
    }),
    group.WithWeight(10),
)

// 创建A/B选择器进行流量分配
abSelector := group.NewABGroupSelector("version-test", "stable", "canary", 0.9)
```

### 区域感知路由

根据用户区域选择就近服务实例：

```go
// 创建区域路由器
regionRouter := group.NewTagRouter("region", "us-east")

// 创建可用区路由器
zoneRouter := group.NewTagRouter("zone", "us-east-1a")

// 创建路由链，优先选择同可用区，其次同区域
chain := group.NewRouterChain(
    zoneRouter,  // 先尝试同可用区
    regionRouter, // 若无可用实例，则选择同区域
)
```

### 自定义选择策略

实现自定义选择器以满足特殊需求：

```go
// 创建按CPU使用率选择的选择器
cpuSelector := selector.NewBaseSelector("cpu-usage", func(ctx context.Context, instances []*naming.Instance) ([]*naming.Instance, error) {
    // 按CPU使用率排序
    sort.Slice(instances, func(i, j int) bool {
        cpuI, _ := strconv.ParseFloat(instances[i].Metadata["cpu_usage"], 64)
        cpuJ, _ := strconv.ParseFloat(instances[j].Metadata["cpu_usage"], 64)
        return cpuI < cpuJ // 选择CPU使用率低的实例
    })
    
    // 返回CPU使用率最低的3个实例
    if len(instances) > 3 {
        return instances[:3], nil
    }
    return instances, nil
})
```
