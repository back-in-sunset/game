# Device Shadow Service

IoT设备影子服务，用于存储和同步设备状态。

## 功能特性

- **GetShadow**: 获取设备影子
- **UpdateShadow**: 更新设备影子（desired状态）
- **ReportState**: 上报设备状态（reported状态）
- **DeleteShadow**: 删除设备影子
- **GetDelta**: 获取状态差异

## 架构

```
device_shadow/
├── rpc/                    # gRPC服务层
│   ├── device_shadow.proto
│   ├── etc/               # 配置文件
│   └── internal/
│       ├── config/        # 配置
│       ├── logic/         # 业务逻辑
│       ├── server/        # gRPC服务器
│       └── svc/           # 服务上下文
├── model/                 # 数据模型
└── Makefile
```

## 快速开始

### 1. 安装依赖

```bash
cd device_shadow/rpc
go mod tidy
```

### 2. 配置

编辑 `rpc/etc/deviceshadow.yaml`：

```yaml
Name: deviceshadow.rpc
ListenOn: 0.0.0.0:8080

Mysql:
  DataSource: root:password@tcp(localhost:3306)/device_shadow?charset=utf8mb4&parseTime=true&loc=Asia/Shanghai

CacheRedis:
  - Host: localhost:6379
    Pass: ""
    Type: node
```

### 3. 创建数据库

```bash
mysql -u root -p < model/device_shadow.sql
```

### 4. 运行服务

```bash
make run
```

### 5. 构建

```bash
make build
```

## API

### gRPC Methods

#### GetShadow
获取设备影子信息

#### UpdateShadow
更新desired状态

#### ReportState
上报reported状态

#### DeleteShadow
删除设备影子

#### GetDelta
获取状态差异

## 开发

### 重新生成Proto代码

```bash
make proto
```

### 运行测试

```bash
make test
```
