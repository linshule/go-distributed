# 分布式日志服务系统

一个用 Go 语言实现的分布式系统入门项目，展示了服务注册、服务发现、健康检查等核心概念。

## 功能特性

- **服务注册中心** - 统一管理所有服务的注册和注销
- **日志服务** - 集中收集和存储日志
- **业务服务** - 图书馆服务示例（图书管理和借阅）
- **服务发现** - 查询已注册的服务列表
- **依赖通知** - 服务变化时自动通知订阅者
- **Web管理界面** - 可视化管理界面
- **健康监控** - 实时监控服务状态

## 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/linshule/go-distributed.git
cd go-distributed
```

### 2. 启动服务

```bash
# 终端1：启动服务注册中心
go run cmd/registryservice/main.go

# 终端2：启动日志服务
go run cmd/logservice/main.go

# 终端3：启动图书馆服务（可选）
go run cmd/libraryservice/main.go

# 终端4：启动Web管理界面（可选）
go run cmd/webservice/main.go

# 终端5：启动监控服务（可选）
go run cmd/monitorservice/main.go
```

### 3. 访问服务

| 服务 | 地址 |
|------|------|
| 服务注册中心 | http://localhost:3000 |
| 日志服务 | http://localhost:4000 |
| 图书馆服务 | http://localhost:5000 |
| Web管理界面 | http://localhost:5002/web |
| 监控服务 | http://localhost:5003/monitor/health |

## 服务依赖关系

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
```

**说明**：图书馆服务依赖于日志服务，在添加书籍和借阅时会自动记录日志。

## 项目结构

```
go-distributed/
├── cmd/                    # 程序入口
│   ├── registryservice/   # 服务注册中心
│   ├── logservice/        # 日志服务
│   ├── libraryservice/    # 图书馆服务
│   ├── providerservice/   # 服务提供者
│   ├── webservice/        # Web界面
│   └── monitorservice/    # 监控服务
├── registry/              # 服务注册模块
├── log/                  # 日志服务模块
├── library/              # 图书馆模块
├── provider/             # 服务提供者模块
├── web/                  # Web界面模块
├── monitor/              # 监控模块
├── service/              # 通用服务模块
└── docs/                 # 详细文档
```

## API 示例

### 注册服务

```bash
curl -X POST http://localhost:3000/services \
  -H "Content-Type: application/json" \
  -d '{"serviceName":"LogService","serviceUrl":"http://localhost:4000"}'
```

### 获取所有服务

```bash
curl http://localhost:3000/services
```

### 发送日志

```bash
curl -X POST -d "Hello World" http://localhost:4000/log
```

### 获取服务健康状态

```bash
curl http://localhost:5003/monitor/health
```

## 学习资源

详细的学习文档请查看 [docs/README.md](docs/README.md)。

## 技术栈

- Go 1.25+
- Go 标准库 (net/http, encoding/json, sync)

## 许可证

MIT License
