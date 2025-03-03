#!/bin/bash
# 启动多个服务器实例，用于演示故障转移功能

echo "Starting server cluster for failover demonstration..."

# 设置基础端口号
BASE_PORT=8001
SERVERS_COUNT=3
PID_FILE="./.server_pids"

# 清理旧的PID文件
[ -f $PID_FILE ] && rm $PID_FILE

# 检查是否有etcd运行
if ! nc -z localhost 2379 2>/dev/null; then
    echo "WARNING: etcd may not be running on port 2379"
    echo "Please make sure etcd is running before continuing"
    sleep 5
fi

# 启动多个服务实例
echo "Starting $SERVERS_COUNT server instances..."

# 创建日志目录
mkdir -p logs

# Server A - 处理100个请求后故障
echo "Starting Server A (Port $BASE_PORT) - Fails after 100 requests for 10s"
cd ../server && go run . -port=$BASE_PORT -id=server-A -fail-after=100 -fail-duration=10s > ../scripts/logs/server_a.log 2>&1 &
echo $! >> $PID_FILE
cd ../scripts

# Server B - 10%概率随机故障
PORT_B=$((BASE_PORT+1))
echo "Starting Server B (Port $PORT_B) - 10% random failure rate"
cd ../server && go run . -port=$PORT_B -id=server-B -fail-rate=0.1 > ../scripts/logs/server_b.log 2>&1 &
echo $! >> $PID_FILE
cd ../scripts

# Server C - 正常运行
PORT_C=$((BASE_PORT+2))
echo "Starting Server C (Port $PORT_C) - Normal operation"
cd ../server && go run . -port=$PORT_C -id=server-C > ../scripts/logs/server_c.log 2>&1 &
echo $! >> $PID_FILE
cd ../scripts

echo "All servers started! Server status:"
echo "  Server A: http://localhost:$BASE_PORT/health"
echo "  Server B: http://localhost:$PORT_B/health"
echo "  Server C: http://localhost:$PORT_C/health"

echo "Starting client demo in 3 seconds..."
sleep 3

# 启动客户端示例
echo "Starting failover client demo..."
cd ../client && go run . > ../scripts/logs/client.log 2>&1 &
echo $! >> $PID_FILE
cd ../scripts

echo ""
echo "To stop all servers, use: ./stop_cluster.sh"
echo "To view logs: tail -f logs/server_*.log or logs/client.log"
echo ""
echo "Press Ctrl+C to exit (servers will keep running in background)..."

# 添加停止集群的脚本
cat > ./stop_cluster.sh << 'EOL'
#!/bin/bash
if [ -f ./.server_pids ]; then
    echo "Stopping all cluster processes..."
    while read pid; do
        if kill -0 $pid 2>/dev/null; then
            echo "Stopping process $pid"
            kill $pid
        fi
    done < ./.server_pids
    rm ./.server_pids
    echo "All processes stopped"
else
    echo "No PID file found, nothing to stop"
fi
EOL
chmod +x ./stop_cluster.sh

# 等待用户输入
read -r -d '' _ </dev/tty