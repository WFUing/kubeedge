#!/bin/bash

# 检查是否传入IP和端口
if [[ $# -ne 2 ]]; then
    echo "Usage: $0 <server_ip> <port>"
    exit 1
fi

server_ip=$1
port=$2

# 定义临时文件夹和文件路径
TMP_DIR="/tmp/benchmark"
mkdir -p "$TMP_DIR"
CPU_RESULT="$TMP_DIR/cpu_result.txt"
MEM_RESULT="$TMP_DIR/mem_result.txt"
FIO_RESULT="$TMP_DIR/fio_result.txt"
IPERF_RESULT="$TMP_DIR/iperf_result.json"

# CPU 测试
cpu_result=$(sysbench cpu --cpu-max-prime=20000 run)
echo "$cpu_result" > "$CPU_RESULT"

# 内存测试
mem_result=$(sysbench memory --memory-total-size=1G run)
echo "$mem_result" > "$MEM_RESULT"

# 磁盘 I/O 测试
fio_result=$(fio --name=randwrite --ioengine=libaio --iodepth=4 --rw=randwrite \
--bs=4k --direct=1 --size=1G --numjobs=4 --runtime=60 --group_reporting --directory=$TMP_DIR)
echo "$fio_result" > "$FIO_RESULT"

# 网络测试（使用 iperf3）
iperf3 -c "$server_ip" -p "$port" -J > "$IPERF_RESULT"

# 解析结果并生成 JSON
cpu_events_per_second=$(grep "events per second:" "$CPU_RESULT" | awk '{print $4}')
mem_ops_per_sec=$(grep "per second" "$MEM_RESULT" | awk -F '[()]' '{print $2}' | awk '{print $1}')
fio_iops=$(grep "IOPS=" "$FIO_RESULT" | awk -F'=' '{print $2}' | awk -F',' '{print $1}')
fio_bw=$(grep "BW=" "$FIO_RESULT" | awk -F'BW=' '{print $2}' | awk -F'KiB/s' '{print $1}')
disk_io_benchmark=$(awk "BEGIN {print $fio_iops * $fio_bw}")

# 解析 iperf3 网络带宽和延迟
latency_ms=$(jq '.end.sum_sent.start.delay_ms' "$IPERF_RESULT")
bandwidth_bps=$(jq '.end.sum_sent.bits_per_second' "$IPERF_RESULT")
bandwidth_mbps=$(awk "BEGIN {print $bandwidth_bps / 1000000}")

# 生成 JSON 格式的结果并通过 echo 输出
cat <<EOF
{
  "cpu_benchmark": $cpu_events_per_second,
  "memory_benchmark": $mem_ops_per_sec,
  "disk_io_benchmark": $disk_io_benchmark,
  "network_benchmark": {
    "latency_ms": $latency_ms,
    "bandwidth_mbps": $bandwidth_mbps
  }
}
EOF

# 清理临时文件夹
rm -rf "$TMP_DIR"
