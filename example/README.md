# example

example目录是fyer-rpc框架的示例目录，包含了一个展示了包括微服务注册发现、负载均衡、故障转移、熔断和监控等功能。目录下简单实现了一个简单的问候服务集群，可以帮助了解如何使用fyer-rpc构建高可用的分布式系统。

## 1.运行示例

运行示例前，需要安装并启动以下服务：

* etcd：服务注册中心
* Prometheus：指标收集和监控

### （1）启动etcd服务（不需要用户名与密码）：
```shell
# 以docker为例
docker run -d -p 2379:2379 quay.io/coreos/etcd:v3.5.0 /usr/local/bin/etcd --listen-client-urls http://0.0.0.0:2379 --advertise-client-urls http://0.0.0.0:2379
```

### （2）启动prometheus服务：

确保Prometheus配置文件（`prometheus.yml`）包含以下内容：

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'fyerrpc'
    static_configs:
      - targets: ['localhost:8081']
```

然后启动prometheus：

```shell
# 以docker为例
docker run -d -p 9090:9090 -v /path/to/prometheus.yml:/etc/prometheus/prometheus.yml prom/prometheus
```

### （3）启动集群

**Windows**：

```shell
cd example/scripts
start_cluster.bat
```

**Linux/macOS**：

```shell
cd example/scripts
chmod +x start_cluster.sh
./start_cluster.sh
```

### （4）预计结果

启动脚本会同时启动三个服务器实例和一个客户端，每个都会在单独的窗口（Windows）或后台（Linux/Mac）运行。

* 服务器实例:

  * 服务器A（端口8001）：处理100个请求后会故障10秒
  * 服务器B（端口8002）：10%概率随机故障
  * 服务器C（端口8003）：正常运行

客户端将自动运行以下测试场景:

* 基本调用测试
* 故障转移测试
* 熔断器测试
* 并发调用测试

### （5）服务监控

#### a.健康检查

可以通过以下 URL 查看各个服务的健康状态：

* 服务器A：http://localhost:18001/health
* 服务器B：http://localhost:18002/health
* 服务器C：http://localhost:18003/health

#### b.Prometheus指标

在Prometheus UI（http://localhost:9090）中，可以查询以下指标：

* fyerrpc_request_total - 请求总数
* fyerrpc_response_time_seconds - 响应时间
* fyerrpc_failover_total - 故障转移总数
* fyerrpc_circuit_breaks - 熔断次数
* fyerrpc_instance_health - 实例健康状态

## 2.项目设计

### （1）总体结构

项目的总体结构如下：

```text
example/
├── client/            - 客户端代码
│   ├── client.go      - 客户端实现
│   ├── config.go      - 配置管理
│   ├── failover.go    - 故障转移实现
│   └── main.go        - 客户端主程序
├── common/            - 公共代码
│   ├── config.go      - 配置定义
│   └── metrics.go     - 指标收集
├── helloworld/        - 服务定义
│   ├── hello.go       - 服务接口
│   └── hello.pb.go    - 服务定义
├── scripts/           - 脚本
│   ├── start_cluster.bat  - Windows启动脚本
│   └── start_cluster.sh   - Linux启动脚本
└── server/            - 服务器代码
    ├── config.go      - 服务器配置
    ├── detector.go    - 健康检测
    ├── main.go        - 服务器主程序
    └── service.go     - 服务实现
```

### （2）总体设计

示例项目的主体由以下部分构成：

* **服务器端**：提供问候服务的多个实例，模拟不同的故障场景。
* **客户端**：连接服务器并展示故障转移功能。
* **注册中心**：使用etcd进行服务注册与发现。
* **监控系统**：使用Prometheus收集和展示指标。

项目还模拟了集中常见的分布式系统故障场景：

1. **计数失败**：服务器A在处理一定数量的请求后会失败一段时间。 
2. **随机失败**：服务器B有一定概率随机失败。 
3. **正常服务**：服务器C始终正常运行。

这些模拟故障帮助展示框架的自动故障转移和熔断功能。

## 3.功能模块

### （1）集群与故障转移（cluster/failover）

故障转移模块提供了在服务实例失败时自动切换到健康实例的功能：

* **故障检测器**：使用TCP探测和自定义健康检查来识别不健康的实例。
* **重试策略**：支持多种重试策略，包括固定间隔、指数退避等。
* **熔断器**：防止对不健康服务的持续请求，减轻系统负担。
* **实例管理**：维护实例健康状态，优先选择健康的实例。

### （2）指标收集（metrics）

指标收集模块提供了全面的系统监控能力：

* **请求计数**：跟踪总请求数、成功请求和失败请求。
* **响应时间**：记录各服务实例的响应时间。
* **故障转移事件**：记录故障转移发生的次数和详情。
* **熔断状态**：监控熔断器的状态变化。

### （3）RPC服务器（rpc/server）

RPC服务器实现了远程过程调用的基本功能：

* **服务注册**：允许将服务实现注册到RPC服务器。
* **请求处理**：接收和处理客户端请求。
* **响应编码解码**：支持多种序列化格式。
* **超时控制**：防止请求处理时间过长。

### （4）服务发现 (discovery)

服务发现模块基于etcd实现了服务注册与发现：

* **服务注册**：服务实例启动时自动注册到etcd。
* **服务发现**：客户端自动发现可用的服务实例。
* **负载均衡**：支持多种负载均衡策略，如轮询、随机、最快响应等。