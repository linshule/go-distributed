# 服务发现功能说明

## 目录

1. [概述](#1-概述)
2. [核心概念](#2-核心概念)
3. [架构设计](#3-架构设计)
4. [API接口](#4-api接口)
5. [使用示例](#5-使用示例)
6. [高级功能](#6-高级功能)

---

## 1. 概述

服务发现（Service Discovery）是分布式系统中让服务能够相互感知和定位的机制。在微服务架构中，服务实例的网络位置是动态变化的，因此需要一种机制来让服务能够发现彼此。

本项目实现了完整的服务发现功能，包括：

- **服务注册**：服务启动时自动向注册中心注册
- **服务注销**：服务停止时自动从注册中心移除
- **服务查询**：按名称、标签等方式查询服务
- **健康检查**：检查服务可用性
- **服务变更通知**：订阅服务变化
- **负载均衡**：从多个健康实例中选择

---

## 2. 核心概念

### 2.1 服务注册信息（Registration）

```go
type Registration struct {
    ServiceName    ServiceName            // 服务名称
    ServiceUrl     string                 // 服务URL
    ServiceVersion string                 // 服务版本
    Metadata       map[string]string      // 元数据
    Tags           []string               // 标签
    HealthCheckURL string                 // 健康检查URL
    RegisteredAt   time.Time             // 注册时间
}
```

**字段说明**：

| 字段 | 说明 | 示例 |
|------|------|------|
| ServiceName | 服务唯一标识 | "LogService" |
| ServiceUrl | 服务访问地址 | "http://localhost:4000" |
| ServiceVersion | 服务版本号 | "1.0.0" |
| Metadata | 自定义元数据 | {"env": "production"} |
| Tags | 服务标签 | ["logging", "core"] |
| HealthCheckURL | 健康检查地址 | "http://localhost:4000" |
| RegisteredAt | 注册时间 | 2024-01-01 10:00:00 |

### 2.2 服务实例（ServiceInstance）

```go
type ServiceInstance struct {
    Name      string            // 服务名称
    URL       string            // 服务URL
    Version   string            // 服务版本
    Metadata  map[string]string // 元数据
    Tags      []string          // 标签
    Healthy   bool              // 健康状态
    Latency   int64             // 响应延迟(毫秒)
    LastCheck time.Time         // 最后检查时间
}
```

---

## 3. 架构设计

### 3.1 组件结构

```
┌─────────────────────────────────────────────────────────┐
│                    服务注册中心 (Registry)                │
│                      端口: 3000                         │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐   │
│  │ 服务注册    │  │ 服务查询    │  │ 健康检查    │   │
│  │ (POST)     │  │ (GET)       │  │ (/health/)  │   │
│  └─────────────┘  └─────────────┘  └─────────────┘   │
└─────────────────────────────────────────────────────────┘
          ▲                    ▲                    ▲
          │                    │                    │
    ┌─────┴─────┐        ┌─────┴─────┐        ┌─────┴─────┐
    │ 服务发现   │        │ 服务发现   │        │ 服务发现   │
    │ (Discovery)│        │ (Discovery)│        │ (Discovery)│
    └─────┬─────┘        └─────┬─────┘        └─────┬─────┘
          │                    │                    │
    ┌─────┴─────┐        ┌─────┴─────┐        ┌─────┴─────┐
    │ 日志服务   │        │ 图书馆服务 │        │ Web服务   │
    └───────────┘        └───────────┘        └───────────┘
```

### 3.2 服务发现流程

```
1. 服务启动
   │
   ▼
2. 向注册中心注册（POST /services）
   │
   ▼
3. 注册中心保存服务信息
   │
   ▼
4. 其他服务查询服务（GET /services/{name}）
   │
   ▼
5. 返回服务URL和元数据
   │
   ▼
6. 直接调用目标服务
```

---

## 4. API接口

### 4.1 服务注册中心 API

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | /services | 注册新服务 |
| DELETE | /services | 注销服务 |
| GET | /services | 获取所有服务 |
| GET | /services/{name} | 按名称查询服务 |
| GET | /services/tag/{tag} | 按标签查询服务 |
| GET | /health | 注册中心健康检查 |
| GET | /health/{name} | 服务健康检查 |

**按名称查询服务示例**：
```bash
# 查询日志服务
curl http://localhost:3000/services/LogService

# 响应
[
  {
    "serviceName": "LogService",
    "serviceUrl": "http://localhost:4000",
    "serviceVersion": "1.0.0",
    "metadata": {"description": "Centralized logging service"},
    "tags": ["logging", "core"],
    "healthCheckUrl": "http://localhost:4000",
    "registeredAt": "2024-01-01T10:00:00Z"
  }
]
```

**按标签查询服务示例**：
```bash
# 查询所有核心服务
curl http://localhost:3000/services/tag/core
```

**服务健康检查示例**：
```bash
# 检查日志服务健康状态
curl http://localhost:3000/health/LogService

# 响应
{
  "serviceName": "LogService",
  "healthy": true,
  "latency": 5
}
```

### 4.2 服务发现 API（端口 5001）

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | /discovery | 获取所有服务（含健康状态） |

**示例**：
```bash
curl http://localhost:5001/discovery

# 响应
{
  "LogService": [
    {
      "name": "LogService",
      "url": "http://localhost:4000",
      "version": "1.0.0",
      "healthy": true,
      "latency": 5
    }
  ],
  "LibraryService": [
    {
      "name": "LibraryService",
      "url": "http://localhost:5000",
      "version": "1.0.0",
      "healthy": true,
      "latency": 10
    }
  ]
}
```

---

## 5. 使用示例

### 5.1 客户端服务发现

```go
import "github.com/linshule/go-distributed/registry"

// 查询所有服务
services, err := registry.GetServices()
if err != nil {
    log.Fatal(err)
}
for _, s := range services {
    fmt.Printf("%s: %s\n", s.ServiceName, s.ServiceUrl)
}

// 按名称查找服务
logService, err := registry.FindService(registry.LogService)
if err != nil {
    log.Fatal(err)
}

// 按标签查找
coreServices, err := registry.FindServicesByTag("core")

// 健康检查
healthy, latency, err := registry.HealthCheck(registry.LogService)
```

### 5.2 高级服务发现

```go
import "github.com/linshule/go-distributed/discovery"

// 创建服务发现实例
d := discovery.New("http://localhost:3000/services")

// 刷新服务列表
d.Refresh()

// 获取所有实例
instances := d.GetInstances("LogService")

// 获取健康实例（带负载均衡）
instance, err := d.GetHealthyInstance("LogService")
if err != nil {
    log.Fatal(err)
}

// 使用实例
fmt.Printf("Calling %s at %s\n", instance.Name, instance.URL)

// 观察服务变化
d.Watch("LogService", func(inst *discovery.ServiceInstance) {
    fmt.Printf("LogService changed: %s\n", inst.URL)
})

// 启动定时刷新
d.StartPolling(10 * time.Second)
```

### 5.3 负载均衡

```go
// 获取一个健康的服务实例（自动负载均衡）
instance, err := discovery.GetHealthyInstance("LogService")
if err != nil {
    // 处理错误
}

// 使用轮询选择健康实例
http.Get(instance.URL + "/your-endpoint")
```

### 5.4 服务变化通知

```go
// 观察特定服务的变化
discovery.Watch("LogService", func(inst *discovery.ServiceInstance) {
    if inst.Healthy {
        fmt.Printf("LogService is now healthy at %s\n", inst.URL)
    } else {
        fmt.Printf("LogService is unhealthy\n")
    }
})
```

---

## 6. 高级功能

### 6.1 缓存机制

服务发现客户端内置缓存，减少对注册中心的请求：

```go
// 设置缓存过期时间（默认30秒）
registry.SetCacheExpiry(60 * time.Second)

// 清除缓存
registry.ClearCache()
```

### 6.2 服务版本管理

支持服务版本控制：

```go
r := registry.Registration{
    ServiceName:    "MyService",
    ServiceUrl:     "http://localhost:8080",
    ServiceVersion: "1.0.0",  // 语义版本
    Metadata: map[string]string{
        "minVersion": "1.0.0",
        "maxVersion": "2.0.0",
    },
}
```

### 6.3 服务标签

使用标签对服务进行分组：

```go
r := registry.Registration{
    ServiceName: "MyService",
    Tags: []string{"api", "v1", "critical"},
}

// 查询特定标签的服务
services, _ := registry.FindServicesByTag("critical")
```

### 6.4 自定义健康检查

```go
r := registry.Registration{
    ServiceName:    "MyService",
    ServiceUrl:     "http://localhost:8080",
    HealthCheckURL: "http://localhost:8080/health",  // 自定义健康检查端点
}
```

---

## 7. 与其他模块的关系

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   日志服务   │────►│  注册中心   │◄────│  服务发现   │
│  (LogService)│     │ (Registry) │     │(Discovery)  │
└─────────────┘     └─────────────┘     └─────────────┘
                           │                   │
                           ▼                   ▼
                    ┌─────────────┐     ┌─────────────┐
                    │  监控服务    │────►│  Web界面    │
                    │ (Monitor)   │     │   (Web)     │
                    └─────────────┘     └─────────────┘
```

- **注册中心**：负责存储服务注册信息
- **服务发现**：提供高级查询和健康检查功能
- **监控服务**：使用服务发现获取服务列表进行检查
- **Web界面**：使用服务发现展示服务状态
