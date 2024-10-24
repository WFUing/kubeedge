package taskexecutor

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type CPUBenchmarkResult struct {
	EventsPerSecond float64 `json:"events_per_second"`
	AvgLatencyMs    float64 `json:"avg_latency_ms"`
}

type MemoryBenchmarkResult struct {
	OperationsPerSecond float64 `json:"operations_per_second"`
	AvgLatencyMs        float64 `json:"avg_latency_ms"`
}

type DiskIOBenchmarkResult struct {
	IOPS          float64 `json:"iops"`
	BandwidthKiBs float64 `json:"bandwidth_kibs"`
	AvgLatencyUs  float64 `json:"avg_latency_us"`
}

type NetWorkBenchmarkResult struct {
	LatencyMs     float64 `json:"latency_ms"`     // 网络时延，单位毫秒
	BandwidthKiBs float64 `json:"bandwidth_kibs"` // 带宽，单位KiB/s
}

type BenchmarkResults struct {
	CPUBenchmark     CPUBenchmarkResult     `json:"cpu_benchmark"`
	MemoryBenchmark  MemoryBenchmarkResult  `json:"memory_benchmark"`
	DiskIOBenchmark  DiskIOBenchmarkResult  `json:"disk_io_benchmark"`
	NetWorkBenchmark NetWorkBenchmarkResult `json:"net_work_benchmark"`
}

func nodeBenchmark() (msg string) {

	var err error
	defer func() {
		if err != nil {
			msg = err.Error()
		}
	}()

	// 收集基准测试结果
	cpuResult, err := cpuBenchmark()
	if err != nil {
		fmt.Printf("CPU benchmark failed: %v\n", err)
		return
	}

	memoryResult, err := memoryBenchmark()
	if err != nil {
		fmt.Printf("Memory benchmark failed: %v\n", err)
		return
	}

	diskResult, err := diskBenchmark()
	if err != nil {
		fmt.Printf("Disk benchmark failed: %v\n", err)
		return
	}

	netWorkResult, err := networkBenchmark()
	if err != nil {
		fmt.Printf("network benchmark failed: %v\n", err)
		return
	}

	// 整合所有的基准测试结果
	results := BenchmarkResults{
		CPUBenchmark:     cpuResult,
		MemoryBenchmark:  memoryResult,
		DiskIOBenchmark:  diskResult,
		NetWorkBenchmark: netWorkResult,
	}

	// 将结果转换为 JSON 并输出
	resultJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal results: %v\n", err)
		return
	}

	msg = string(resultJSON)
	return msg
}

func cpuBenchmark() (CPUBenchmarkResult, error) {
	start := time.Now()

	// 执行 1 秒的 CPU 密集型任务，测量事件数
	events := 0
	for time.Since(start) < time.Second {
		for i := 0; i < 1000000; i++ {
			_ = i * i
		}
		events++
	}

	// 计算每秒事件数
	duration := time.Since(start).Seconds()
	eventsPerSecond := float64(events) / duration

	// 计算平均延迟
	avgLatencyMs := (duration / float64(events)) * 1000

	// 返回 CPU 基准测试结果
	return CPUBenchmarkResult{
		EventsPerSecond: eventsPerSecond,
		AvgLatencyMs:    avgLatencyMs,
	}, nil
}

func memoryBenchmark() (MemoryBenchmarkResult, error) {
	start := time.Now()

	operations := 0
	for time.Since(start) < time.Second {
		_ = make([]byte, 1024*1024) // 每次分配 1 MB 的内存
		operations++
	}

	// 计算每秒操作数
	duration := time.Since(start).Seconds()
	operationsPerSecond := float64(operations) / duration

	// 计算平均延迟
	avgLatencyMs := (duration / float64(operations)) * 1000

	// 返回内存基准测试结果
	return MemoryBenchmarkResult{
		OperationsPerSecond: operationsPerSecond,
		AvgLatencyMs:        avgLatencyMs,
	}, nil
}

func diskBenchmark() (DiskIOBenchmarkResult, error) {
	// 磁盘写入 1MB 文件
	data := make([]byte, 4096) // 每次写入 4KB 数据块
	file, err := os.Create("tempfile")
	if err != nil {
		return DiskIOBenchmarkResult{}, err
	}
	defer file.Close()

	start := time.Now()
	iops := 0
	totalBytes := 0

	for time.Since(start) < time.Second {
		_, err := file.Write(data)
		if err != nil {
			return DiskIOBenchmarkResult{}, err
		}
		iops++
		totalBytes += len(data)
	}

	// 计算 IOPS 和带宽
	duration := time.Since(start).Seconds()
	iopsPerSecond := float64(iops) / duration
	bandwidthKiB := float64(totalBytes) / 1024 / duration

	// 假设延迟为 1000 微秒
	avgLatencyUs := (duration / float64(iops)) * 1e6

	// 返回磁盘基准测试结果
	return DiskIOBenchmarkResult{
		IOPS:          iopsPerSecond,
		BandwidthKiBs: bandwidthKiB,
		AvgLatencyUs:  avgLatencyUs,
	}, nil
}

func networkBenchmark() (NetWorkBenchmarkResult, error) {
	start := time.Now()

	// Simulating network latency measurement (for demonstration, assume 20ms)
	latencyMs := 20.0

	// Simulating network bandwidth measurement (sending dummy data)
	totalBytesSent := 1024 * 1024 * 10 // Simulating 10MB of data sent
	duration := time.Since(start).Seconds()

	// Calculating bandwidth in KiB/s
	bandwidthKiB := float64(totalBytesSent) / 1024 / duration

	// Returning the network benchmark results
	return NetWorkBenchmarkResult{
		LatencyMs:     latencyMs,
		BandwidthKiBs: bandwidthKiB,
	}, nil
}
