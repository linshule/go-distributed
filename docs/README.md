# 分布式日志服务系统 - 零基础入门指南

## 目录

1. [项目概述](#1-项目概述)
2. [前置知识](#2-前置知识)
3. [项目结构](#3-项目结构)
4. [核心概念](#4-核心概念)
5. [快速开始](#5-快速开始)
6. [代码详解](#6-代码详解)
7. [API接口说明](#7-api接口说明)
8. [运行流程图](#8-运行流程图)
9. [常见问题](#9-常见问题)

---

## 1. 项目概述

### 1.1 什么是分布式日志服务？

想象一下：一个大型网站有多个服务器（Web服务器、数据库服务器、文件服务器等），每个服务器都会产生日志。如果这些日志分散在各处，排查问题时会非常麻烦。

这个项目实现了一个**分布式日志服务系统**，它可以：
- **收集日志**：各个服务把日志发送到统一的地方
- **服务发现**：通过注册中心知道哪些服务正在运行
- **统一存储**：把所有日志写入同一个文件
- **服务监控**：监控各个服务的健康状态
- **依赖通知**：当依赖的服务发生变化时通知调用方

### 1.2 项目组成

本项目包含以下服务：

| 服务 | 端口 | 功能 |
|------|------|------|
| 服务注册中心 | 3000 | 记录哪些服务正在运行 |
| 日志服务 | 4000 | 接收并保存日志 |
| 图书馆服务 | 5000 | 图书管理和借阅服务 |
| 服务提供者 | 5001 | 服务发现和依赖通知 |
| Web管理界面 | 5002 | 可视化管理界面 |
| 监控服务 | 5003 | 服务健康状态监控 |

### 1.3 服务依赖关系

本项目展示了分布式系统中服务之间的依赖关系：

```
┌─────────────────┐
│  服务注册中心    │  (端口 3000)
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌───────┐  ┌───────────┐
│日志服务│  │图书馆服务 │
│(4000) │◄─┤  (5000)  │ 依赖日志服务
└───────┘  └───────────┘
    ▲
    │
┌───┴────┐
│Web界面  │
│(5002)  │
└────────┘
    ▲
    │
┌───┴──────┐
│ 监控服务  │
│ (5003)   │
└──────────┘
```

**说明**：
- **图书馆服务** 依赖 **日志服务**：图书馆服务在以下情况会记录日志
  - 服务启动时
  - 添加新书籍时
  - 借阅书籍时
- **Web界面** 依赖 **注册中心、日志服务、图书馆服务**
- **监控服务** 依赖 **注册中心**

---

## 2. 前置知识

### 2.1 需要安装的工具

1. **Go语言** (1.25.4或更高版本)
   - 下载地址：https://go.dev/dl/
   - 安装后打开终端，输入 `go version` 验证

2. **Git** (可选)
   - 用于版本控制

### 2.2 了解HTTP协议

项目中用到了HTTP的几种方法：

| 方法 | 含义 | 用途 |
|------|------|------|
| GET | 获取数据 | 查看服务列表 |
| POST | 提交数据 | 注册服务、发送日志 |
| DELETE | 删除数据 | 注销服务 |

### 2.3 了解JSON

JSON是一种常用的数据格式，用于前后端数据交换。

```json
{
  "serviceName": "LogService",
  "serviceUrl": "http://localhost:4000"
}
```

---

## 3. 项目结构

```
go-distributed/                          # 项目根目录
├── go.mod                         # Go模块配置文件
├── cmd/                           # 程序入口目录
│   ├── registryservice/
│   │   └── main.go               # 服务注册中心入口
│   ├── logservice/
│   │   └── main.go               # 日志服务入口
│   ├── libraryservice/
│   │   └── main.go               # 图书馆服务入口
│   ├── providerservice/
│   │   └── main.go               # 服务提供者入口
│   ├── webservice/
│   │   └── main.go               # Web管理界面入口
│   └── monitorservice/
│       └── main.go               # 监控服务入口
├── registry/                      # 服务注册模块
│   ├── registration.go           # 服务注册数据结构
│   ├── server.go                 # 注册中心服务端
│   └── client.go                 # 注册中心客户端
├── log/                           # 日志服务模块
│   └── server.go                 # 日志服务实现
├── library/                       # 图书馆服务模块
│   └── server.go                 # 图书管理和借阅
├── provider/                      # 服务提供者模块
│   └── provider.go               # 服务发现和依赖通知
├── web/                           # Web界面模块
│   └── server.go                 # 可视化管理界面
├── monitor/                       # 监控服务模块
│   └── monitor.go                # 服务健康检查
├── service/                       # 通用服务模块
│   └── service.go                # 服务启动辅助函数
└── docs/                          # 文档目录
    └── README.md                  # 本文档
```

### 3.1 各文件作用

| 文件 | 作用 |
|------|------|
| `go.mod` | 定义项目名称和依赖 |
| `cmd/registryservice/main.go` | 启动服务注册中心 |
| `cmd/logservice/main.go` | 启动日志服务 |
| `cmd/libraryservice/main.go` | 启动图书馆服务 |
| `cmd/providerservice/main.go` | 启动服务提供者 |
| `cmd/webservice/main.go` | 启动Web管理界面 |
| `cmd/monitorservice/main.go` | 启动监控服务 |
| `registry/registration.go` | 定义服务注册的数据结构 |
| `registry/server.go` | 实现服务注册中心的核心逻辑 |
| `registry/client.go` | 供其他服务调用注册中心的工具 |
| `log/server.go` | 实现日志服务的核心逻辑 |
| `library/server.go` | 实现图书馆服务和借阅管理 |
| `provider/provider.go` | 实现服务发现和依赖通知 |
| `web/server.go` | 实现Web管理界面 |
| `monitor/monitor.go` | 实现服务健康监控 |
| `service/service.go` | 封装服务启动的通用代码 |

---

## 4. 核心概念

### 4.1 服务注册中心 (Registry Service)

**作用**：相当于"公司前台"，记录所有在线的服务。

- 启动时监听 3000 端口
- 维护一份"服务清单"
- 接收服务注册（POST）和注销（DELETE）请求

### 4.2 服务注册 (Registration)

每个服务启动时需要向注册中心"报到"，提交以下信息：

```go
type Registration struct {
    ServiceName string   // 服务名称，如 "LogService"
    ServiceUrl  string   // 服务地址，如 "http://localhost:4000"
}
```

### 4.3 日志服务 (Log Service)

**作用**：专门负责记录日志的服务。

- 启动时监听 4000 端口
- 提供 `/log` 接口接收日志
- 将日志写入 `distributed.log` 文件

### 4.4 服务依赖 (Service Dependencies)

在分布式系统中，服务之间往往存在依赖关系。本项目展示了以下依赖：

**图书馆服务依赖日志服务**：
- 图书馆服务启动时，会向日志服务发送启动日志
- 添加新书籍时，会记录书籍信息到日志服务
- 借阅书籍时，会记录借阅信息到日志服务

**为什么要记录日志？**
- 追踪系统行为
- 排查问题
- 审计和监控

---

## 5. 快速开始

### 5.1 克隆项目

```bash
git clone <仓库地址>
cd go-distributed
```

### 5.2 启动服务注册中心

打开一个终端，运行：

```bash
go run cmd/registryservice/main.go
```

应该看到类似输出：

```
Registry Service started on :3000
```

### 5.3 启动日志服务

打开另一个终端，运行：

```bash
go run cmd/logservice/main.go
```

应该看到类似输出：

```
Log Service started on :4000
Log Service registered successfully
```

### 5.4 测试日志服务

再打开一个终端，发送一条日志：

```bash
curl -X POST -d "Hello from test" http://localhost:4000/log
```

### 5.5 查看日志文件

检查是否生成了日志文件：

```bash
# Windows
type distributed.log

# Linux/Mac
cat distributed.log
```

### 5.6 启动图书馆服务（可选）

打开另一个终端，运行：

```bash
go run cmd/libraryservice/main.go
```

图书馆服务提供图书管理和借阅功能，预设了3本示例图书。

### 5.7 启动Web管理界面（可选）

```bash
go run cmd/webservice/main.go
```

启动后可以在浏览器访问 http://localhost:5002/web 查看管理界面。

### 5.8 启动监控服务（可选）

```bash
go run cmd/monitorservice/main.go
```

监控服务会定期检查所有已注册服务的健康状态。

---

## 6. 代码详解

### 6.1 go.mod - 项目配置

```go
module github.com/linshule/go-distributed

go 1.25.4
```

- `module`：项目名称
- `go`：Go版本要求

### 6.2 registry/registration.go - 数据结构

```go
// ServiceName 是服务名称的类型
type ServiceName string

// 定义具体的服务名称常量
const (
    LogService ServiceName = "LogService"
)

// Registration 服务注册信息
type Registration struct {
    ServiceName ServiceName  // 服务名称
    ServiceUrl  string       // 服务URL
}
```

**解释**：
- 定义了一个`ServiceName`类型，实际上就是字符串
- `LogService = "LogService"` 是预定义的服务名

### 6.3 registry/server.go - 注册中心服务端

```go
type RegistryServer struct {
    services map[ServiceName]string  // 服务列表
    mu       sync.Mutex              // 保护并发访问
}

func (r *RegistryServer) addService(registration Registration) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.services[registration.ServiceName] = registration.ServiceUrl
}

func (r *RegistryServer) removeService(serviceName ServiceName) {
    r.mu.Lock()
    defer r.mu.Unlock()
    delete(r.services, serviceName)
}
```

**关键点**：
- 使用`map`存储服务信息，key是服务名，value是URL
- 使用`sync.Mutex`（互斥锁）保证线程安全
- `defer r.mu.Unlock()` 确保函数结束后解锁

### 6.4 registry/client.go - 注册中心客户端

```go
// RegistrationService 向注册中心注册服务
func RegistrationService(r registry.Registration) error {
    jsonData, _ := json.Marshal(r)
    // 发送POST请求到注册中心
    http.Post("http://localhost:3000/services",
              "application/json",
              bytes.NewBuffer(jsonData))
    return nil
}

// ShutdownService 通知注册中心服务关闭
func ShutdownService(url string) error {
    // 发送DELETE请求到注册中心
    client := &http.Client{}
    req, _ := http.NewRequest("DELETE",
        "http://localhost:3000/services",
        nil)
    client.Do(req)
    return nil
}
```

**解释**：
- `RegistrationService`：服务启动时调用，把服务信息发给注册中心
- `ShutdownService`：服务关闭时调用，通知注册中心移除该服务

### 6.5 log/server.go - 日志服务

```go
type LogServer struct {
    File *os.File  // 日志文件
}

// Run 初始化日志服务
func (l *LogServer) Run(destination string) error {
    // 以追加模式打开文件，不存在则创建
    file, err := os.OpenFile(destination,
        os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
    l.File = file
    return err
}

// HandleLog 处理日志请求
func (l *LogServer) HandleLog(w http.ResponseWriter,
                               r *http.Request) {
    // 读取请求体
    body, _ := io.ReadAll(r.Body)
    // 写入文件
    l.File.Write(body)
    l.File.WriteString("\n")
}
```

**关键点**：
- `os.OpenFile` 的 flags：
  - `O_CREATE`：文件不存在则创建
  - `O_WRONLY`：只写模式
  - `O_APPEND`：追加模式（不清除原有内容）
- 权限 `0600`：只有所有者可读写

### 6.6 service/service.go - 通用服务启动

```go
// Start 启动服务并注册
func Start(ctx context.Context, host, port string,
           reg registry.Registration,
           registerHandlersFunc func()) {

    // 1. 注册HTTP处理器
    registerHandlersFunc()

    // 2. 启动HTTP服务器
    addr := host + ":" + port
    server := &http.Server{Addr: addr}
    go server.ListenAndServe()

    // 3. 注册到服务中心
    registry.RegistrationService(reg)
}
```

**作用**：封装了所有服务都要做的重复工作：
- 注册HTTP路由
- 启动HTTP服务器
- 向注册中心注册服务

### 6.7 cmd/logservice/main.go - 日志服务入口

```go
func main() {
    // 1. 创建日志服务实例
    logServer := &log.LogServer{}

    // 2. 初始化（打开日志文件）
    logServer.Run("./distributed.log")

    // 3. 启动服务
    service.Start(
        context.Background(),
        "localhost",                    // 主机地址
        "4000",                         // 端口
        registry.Registration{          // 注册信息
            ServiceName: registry.LogService,
            ServiceUrl:  "http://localhost:4000",
        },
        logServer.RegisterHandlers,     // 注册路由
    )
}
```

### 6.8 library/server.go - 图书馆服务

```go
// Book 书籍结构
type Book struct {
    ID     string `json:"id"`
    Title  string `json:"title"`
    Author string `json:"author"`
}

// BorrowRecord 借阅记录
type BorrowRecord struct {
    BookID   string `json:"book_id"`
    Borrower string `json:"borrower"`
}
```

**关键功能**：
- `addBook`：添加书籍到图书馆
- `listBooks`：列出所有书籍
- `borrowBook`：借阅书籍
- `getBorrowRecords`：获取借阅记录

### 6.9 provider/provider.go - 服务提供者

```go
// ServiceProvider 服务提供者
type ServiceProvider struct {
    registrations []registry.Registration
    notifyLock    sync.RWMutex
    notifyMap     map[string][]chan<- registry.Registration
}
```

**关键功能**：
- `Subscribe`：订阅服务变化通知
- `FindService`：查找指定服务
- `GetServices`：获取所有服务
- 当服务注册或注销时，自动通知订阅者

### 6.10 web/server.go - Web管理界面

Web界面使用HTML+JavaScript实现，提供以下功能：
- 查看已注册服务列表
- 发送日志到日志服务
- 查看图书馆书籍和借阅记录

### 6.11 library/server.go - 图书馆服务（服务依赖示例）

```go
// LogServiceURL 日志服务地址
var LogServiceURL = "http://localhost:4000/log"

// sendLog 发送日志到日志服务
func sendLog(message string) {
    buf := bytes.NewBuffer([]byte(message))
    resp, err := http.Post(LogServiceURL, "text/plain", buf)
    if err != nil {
        log.Printf("Failed to send log: %v\n", err)
        return
    }
    defer resp.Body.Close()
}
```

**服务依赖说明**：
- 图书馆服务通过HTTP调用日志服务的 `/log` 接口来记录日志
- 在 `init()` 函数中，服务启动时发送启动日志
- 在 `addBook()` 函数中，添加书籍时发送日志
- 在 `borrowBook()` 函数中，借阅书籍时发送日志

### 6.11 monitor/monitor.go - 监控服务

```go
// ServiceStatus 服务状态
type ServiceStatus struct {
    Name      string    `json:"name"`
    URL       string    `json:"url"`
    Status    string    `json:"status"`     // "healthy", "unhealthy", "unknown"
    LastCheck time.Time `json:"last_check"` // 最后检查时间
    Latency   int64     `json:"latency"`    // 响应延迟（毫秒）
}
```

**关键功能**：
- 定期检查所有已注册服务的可用性
- 记录服务响应延迟
- 提供HTTP接口查询服务健康状态

---

## 7. API接口说明

### 7.1 服务注册中心 API

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | /services | 注册新服务 |
| DELETE | /services | 注销服务 |

**注册服务请求体**：
```json
{
  "serviceName": "LogService",
  "serviceUrl": "http://localhost:4000"
}
```

### 7.2 日志服务 API

| 方法 | 路径 | 功能 |
|------|------|------|
| POST | /log | 写入日志 |

**请求体**：纯文本内容

**示例**：
```bash
# 写入日志
curl -X POST -d "这是测试日志" http://localhost:4000/log
```

### 7.3 图书馆服务 API

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | /library/books | 获取所有书籍 |
| GET | /library/book/{id} | 获取指定书籍 |
| POST | /library/books | 添加新书籍 |
| POST | /library/borrow | 借阅书籍 |
| GET | /library/borrow | 获取借阅记录 |

**注意**：图书馆服务依赖于日志服务。上述操作会自动记录日志到日志服务。

**示例**：
```bash
# 获取所有书籍
curl http://localhost:5000/library/books

# 借阅书籍
curl -X POST -H "Content-Type: application/json" \
  -d '{"book_id":"1","borrower":"张三"}' \
  http://localhost:5000/library/borrow
```

### 7.4 服务注册中心 API（扩展）

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | /services | 获取所有已注册服务 |

**示例**：
```bash
# 获取所有服务
curl http://localhost:3000/services
```

### 7.5 监控服务 API

| 方法 | 路径 | 功能 |
|------|------|------|
| GET | /monitor/health | 获取所有服务健康状态 |
| GET | /monitor/health/{服务名} | 获取指定服务健康状态 |

**示例**：
```bash
# 获取所有服务健康状态
curl http://localhost:5003/monitor/health

# 获取日志服务健康状态
curl http://localhost:5003/monitor/health/LogService
```

---

## 8. 运行流程图

```
┌─────────────────────────────────────────────────────────────┐
│                      启动服务注册中心                         │
│                       (端口 3000)                            │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      启动日志服务                             │
│                       (端口 4000)                            │
│                                                             │
│   1. 创建 LogServer 实例                                    │
│   2. 打开/创建日志文件                                       │
│   3. 启动 HTTP 服务器                                        │
│   4. 向注册中心 POST 服务信息  ──────────────────────────────┼──► 注册中心记录
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      客户端发送日志                           │
│                                                             │
│   curl -X POST -d "日志内容" http://localhost:4000/log      │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      日志服务处理                             │
│                                                             │
│   1. 接收 HTTP POST 请求                                     │
│   2. 读取请求体（日志内容）                                   │
│   3. 写入 distributed.log 文件                             │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      服务关闭时                               │
│                                                             │
│   1. 监听系统信号（如 Ctrl+C）                               │
│   2. 向注册中心发送 DELETE 请求                              │
│   3. 注册中心移除服务记录                                     │
└─────────────────────────────────────────────────────────────┘
```

---

## 9. 常见问题

### Q1: 启动时提示 "port already in use"

端口被占用了。检查是否有其他程序占用了 3000 或 4000 端口。

**解决方法**：
```bash
# Windows 查看端口占用
netstat -ano | findstr "3000"

# 结束占用进程
taskkill /PID <进程ID> /F
```

### Q2: 日志文件没有生成

检查当前目录下是否有写权限。

**解决方法**：
```bash
# 给当前目录添加写权限 (Linux/Mac)
chmod 777 .
```

### Q3: 服务注册失败

检查注册中心是否已启动。

**排查步骤**：
1. 确认注册中心已运行：`curl http://localhost:3000/services`
2. 检查端口是否正确

### Q4: 如何添加新的服务？

1. 在 `registry/registration.go` 添加新的服务名常量
2. 在 `cmd/` 下创建新的服务目录
3. 参考日志服务的启动方式编写服务代码

### Q5: 这个项目可以用到生产环境吗？

**不建议**。这是学习项目，生产环境需要：
- HTTPS 支持
- 认证授权
- 集群高可用
- 持久化存储
- 监控告警

---

## 10. 扩展学习

### 10.1 后续可以添加的功能

1. **健康检查**：定时检查服务是否存活
2. **服务发现**：让服务可以查询其他服务的地址
3. **负载均衡**：多个日志服务轮询处理
4. **日志轮转**：按日期或大小切割日志文件

### 10.2 相关技术

- **Consul/Etcd/ZooKeeper**：成熟的服务注册中心
- **ELK Stack**：企业级日志解决方案
- **gRPC**：高性能 RPC 框架

---

## 参考资料

- Go 语言官方文档：https://go.dev/doc/
- Go 标准库 - net/http：https://pkg.go.dev/net/http
- RESTful API 设计：https://restfulapi.net/
