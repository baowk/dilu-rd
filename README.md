# dilu-rd

Go 语言服务注册与发现组件，支持 **etcd** 和 **Consul** 两种注册中心，适用于 HTTP 和 gRPC 服务。

[dilu](https://github.com/baowk/dilu) 微服务生态的核心组件，与 [dilu-core](https://github.com/baowk/dilu-core)、[dilu-gateway](https://github.com/baowk/dilu-gateway) 无缝集成。

## 特性

- 支持 etcd / Consul 两种注册中心
- 支持 HTTP / gRPC 协议的服务注册
- 心跳续约与自动摘除
- 多种负载均衡算法（轮询、随机）
- 内置 dilu-core 适配器，3 行代码接入微服务
- 可独立使用，不依赖 dilu 框架

## 安装

```bash
go get github.com/baowk/dilu-rd@latest
```

## 快速开始

### 方式一：配合 dilu-core（推荐）

dilu-core v1.2.0+ 已内置 `/health` 端点，配合适配器即可自动注册：

```go
import (
    "github.com/baowk/dilu-core/core"
    "github.com/baowk/dilu-rd/adapter"
)

// 创建注册中心客户端
cfg := adapter.DefaultConfig("consul", []string{"localhost:8500"})
registry, err := adapter.NewDiluRegistry(cfg)
if err != nil {
    log.Fatal(err)
}

// 注入 dilu-core，启动时自动注册，关闭时自动注销
core.GetApp().SetRegistry(registry)
```

### 方式二：独立使用

```go
import (
    "github.com/baowk/dilu-rd/config"
    "github.com/baowk/dilu-rd/rd"
)

cfg := &config.Config{
    Enable:    true,
    Driver:    "consul", // 或 "etcd"
    Endpoints: []string{"localhost:8500"},
    Scheme:    "http",
    Timeout:   10 * time.Second,
    Registers: []*config.RegisterNode{
        {
            Name:        "my-service",
            Addr:        "192.168.1.100",
            Port:        8080,
            Protocol:    "http",
            HealthCheck: "http://192.168.1.100:8080/health",
            Tags:        []string{"dev"},
        },
    },
    Discoveries: []*config.DiscoveryNode{
        {
            Enable:              true,
            Name:                "my-service",
            SchedulingAlgorithm: "robin", // robin | random
        },
    },
}

client, err := rd.NewRDClient(cfg)
if err != nil {
    log.Fatal(err)
}
defer client.Deregister()

// 获取可用服务节点
node, err := client.GetService("my-service", clientIP)
fmt.Println(node.GetUrl()) // http://192.168.1.100:8080
```

## 连接字符串

| 注册中心 | Endpoints 示例 |
|---------|---------------|
| Consul  | `["localhost:8500"]` |
| etcd    | `["localhost:2379", "localhost:2380"]` |

## 负载均衡

| 算法 | 配置值 | 说明 |
|------|--------|------|
| 轮询 | `robin` | 依次分配请求（默认） |
| 随机 | `random` | 随机选择节点 |

## 架构

```
dilu-rd
├── config/       # 配置结构
├── driver/
│   ├── etcd/     # etcd 驱动
│   └── consul/   # Consul 驱动
├── models/       # ServiceNode 模型
├── rd/           # RDClient 接口与工厂
├── scheduling/   # 负载均衡算法
│   └── impl/     # robin / random
├── adapter/      # dilu-core 适配器
└── examples/     # 使用示例
```

## 生态

| 组件 | 说明 |
|------|------|
| [dilu-core](https://github.com/baowk/dilu-core) | 应用框架，提供 `/health` `/metrics` `/routes` 端点 |
| [dilu-gateway](https://github.com/baowk/dilu-gateway) | API 网关，支持从注册中心动态发现上游 |
| [dilu](https://github.com/baowk/dilu) | 基于 Gin+GORM 的全栈脚手架 |

## License

MIT
