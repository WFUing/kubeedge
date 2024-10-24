#!/bin/bash

# 参数校验
if [[ $# -ne 2 ]]; then
    echo "Usage: $0 <server_ip> <port>"
    exit 1
fi

server_ip=$1
port=$2

# 创建临时文件夹
TMP_DIR="./tmp"
mkdir -p "$TMP_DIR"

# 定义临时文件路径
CPU_RESULT="$TMP_DIR/cpu_result.txt"
MEM_RESULT="$TMP_DIR/mem_result.txt"
FIO_RESULT="$TMP_DIR/fio_result.txt"
IPERF_RESULT="$TMP_DIR/iperf3_result.json"

# CPU 测试
echo "Running CPU benchmark..."
cpu_result=$(sysbench cpu --cpu-max-prime=20000 run)
echo "$cpu_result" > "$CPU_RESULT"

# 内存测试
echo "Running memory benchmark..."
mem_result=$(sysbench memory --memory-total-size=1G run)
echo "$mem_result" > "$MEM_RESULT"

# 磁盘 I/O 测试
echo "Running disk I/O benchmark..."
fio_result=$(fio --name=randwrite --ioengine=libaio --iodepth=4 --rw=randwrite \
--bs=4k --direct=1 --size=1G --numjobs=4 --runtime=60 --group_reporting --directory="$TMP_DIR")
echo "$fio_result" > "$FIO_RESULT"

# 网络带宽测试
echo "Running network bandwidth benchmark..."
iperf3_result=$(iperf3 -c "$server_ip" -p "$port" -J)
echo "$iperf3_result" > "$IPERF_RESULT"

# 解析结果并生成 JSON
echo "Parsing results and generating JSON..."

# 解析 CPU 测试结果
cpu_events_per_second=$(grep "events per second:" "$CPU_RESULT" | awk '{print $4}')
cpu_latency_avg=$(grep "avg:" "$CPU_RESULT" | awk '{print $2}' | sed 's/ms//')
cpu_latency_95th=$(grep "95th percentile:" "$CPU_RESULT" | awk '{print $3}' | sed 's/ms//')

# 解析内存测试结果
mem_ops_per_sec=$(grep "per second" "$MEM_RESULT" | awk -F '[()]' '{print $2}' | awk '{print $1}')
mem_throughput=$(grep "transferred (" "$MEM_RESULT" | awk '{print $5}' | sed 's/MiB\/sec)//')
mem_latency_avg=$(grep "avg:" "$MEM_RESULT" | awk '{print $2}' | sed 's/ms//')

# 解析磁盘 I/O 测试结果
fio_iops=$(grep "IOPS=" "$FIO_RESULT" | awk -F'=' '{print $2}' | awk -F',' '{print $1}')
fio_bw=$(grep "BW=" "$FIO_RESULT" | awk -F'BW=' '{print $2}' | awk -F'KiB/s' '{print $1}')
fio_lat_avg=$(grep " lat (usec):" -A 1 "$FIO_RESULT" | awk -F '[,,]' '{print $3}' | awk '{print $1}' | sed 's/.*avg=\([0-9\.]*\).*/\1/')

# 解析网络带宽测试结果
latency_ms=$(jq '.end.sum_sent.start.delay_ms' "$IPERF_RESULT")
bandwidth_mbps=$(jq '.end.sum_sent.bits_per_second' "$IPERF_RESULT")
bandwidth_mbps=$(echo "scale=2; $bandwidth_mbps / 1000000" | bc)

# 定义一个函数，用于有条件地写入键值对
write_json_field() {
  local key="$1"
  local value="$2"
  local comma="$3"

  if [[ -n "$value" ]]; then
    echo "    \"$key\": $value$comma" >> benchmark_result.json
  fi
}

# 生成 JSON 格式的结果
echo "{" > benchmark_result.json

# 写入 CPU 基准测试结果
echo '  "cpu_benchmark": {' >> benchmark_result.json
write_json_field "events_per_second" "$cpu_events_per_second" ","
write_json_field "avg_latency_ms" "$cpu_latency_avg" ","
write_json_field "95th_percentile_latency_ms" "$cpu_latency_95th" ""
echo '  },' >> benchmark_result.json

# 写入内存基准测试结果
echo '  "memory_benchmark": {' >> benchmark_result.json
write_json_field "operations_per_second" "$mem_ops_per_sec" ","
write_json_field "throughput_mibs" "$mem_throughput" ","
write_json_field "avg_latency_ms" "$mem_latency_avg" ""
echo '  },' >> benchmark_result.json

# 写入磁盘 I/O 基准测试结果
echo '  "disk_io_benchmark": {' >> benchmark_result.json
write_json_field "iops" "$fio_iops" ","
write_json_field "bandwidth_kibs" "$fio_bw" ","
write_json_field "avg_latency_us" "$fio_lat_avg"
echo '  },' >> benchmark_result.json

# 写入网络基准测试结果
echo '  "network_benchmark": {' >> benchmark_result.json
write_json_field "latency_ms" "$latency_ms" ","
write_json_field "bandwidth_mbps" "$bandwidth_mbps" ""
echo '  }' >> benchmark_result.json

# 关闭 JSON 对象
echo '}' >> benchmark_result.json

echo "基准测试结果已保存到 benchmark_result.json"

# 打印 JSON 文件内容
cat benchmark_result.json

# 清理临时文件夹
#rm -rf "$TMP_DIR"