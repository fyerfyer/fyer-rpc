@echo off
REM 启动多个服务器实例，用于演示故障转移功能
echo Starting server cluster for failover demonstration...

REM 设置基础端口号
set BASE_PORT=8001
set SERVERS_COUNT=3

REM 检查是否有etcd运行
netstat -ano | findstr ":2379" > nul
if %ERRORLEVEL% NEQ 0 (
    echo WARNING: etcd may not be running on port 2379
    echo Please make sure etcd is running before continuing
    timeout /t 5
)

REM 启动多个服务实例
echo Starting %SERVERS_COUNT% server instances...

REM Server A - 处理100个请求后故障
start "Server A (Port %BASE_PORT%)" cmd /c "cd ..\server && go run . -port=%BASE_PORT% -id=server-A -fail-after=100 -fail-duration=10s"

REM Server B - 10%概率随机故障
set /a "PORT_B=%BASE_PORT%+1"
start "Server B (Port %PORT_B%)" cmd /c "cd ..\server && go run . -port=%PORT_B% -id=server-B -fail-rate=0.1"

REM Server C - 正常运行
set /a "PORT_C=%BASE_PORT%+2"
start "Server C (Port %PORT_C%)" cmd /c "cd ..\server && go run . -port=%PORT_C% -id=server-C"

echo All servers started! Server status:
echo   Server A: http://localhost:%BASE_PORT%/health
echo   Server B: http://localhost:%PORT_B%/health
echo   Server C: http://localhost:%PORT_C%/health

echo Starting client demo in 3 seconds...
timeout /t 3 > nul

REM 启动客户端示例
start "Failover Client Demo" cmd /c "cd ..\client && go run ."

echo.
echo To stop all servers, close the command windows or run stop_cluster.bat
echo.
echo Press Ctrl+C to exit this console (servers will keep running)...
cmd /k