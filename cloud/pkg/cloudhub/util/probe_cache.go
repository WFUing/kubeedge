package util

// Probe 结构体定义
type Probe struct {
	CPU     CPUBenchmark     `json:"cpu_benchmark"`
	Memory  MemoryBenchmark  `json:"memory_benchmark"`
	Disk    DiskIOBenchmark  `json:"disk_io_benchmark"`
	NetWork NetWorkBenchmark `json:"net_work_benchmark"`
}

// CPUBenchmark 表示 CPU 的基准测试结果
type CPUBenchmark struct {
	EventsPerSecond float64 `json:"events_per_second"` // 每秒事件数
	AvgLatencyMs    float64 `json:"avg_latency_ms"`    // 平均延迟，单位毫秒
}

// MemoryBenchmark 表示内存的基准测试结果
type MemoryBenchmark struct {
	OperationsPerSecond float64 `json:"operations_per_second"` // 每秒操作数
	AvgLatencyMs        float64 `json:"avg_latency_ms"`        // 平均延迟，单位毫秒
}

// DiskIOBenchmark 表示磁盘 I/O 的基准测试结果
type DiskIOBenchmark struct {
	IOPS          int     `json:"iops"`           // IOPS（每秒输入输出操作数）
	BandwidthKiBs float64 `json:"bandwidth_kibs"` // 带宽，单位KiB/s
	AvgLatencyUs  float64 `json:"avg_latency_us"` // 平均延迟，单位微秒
}

// NetWorkBenchmark 表示网络性能的基准测试结果
type NetWorkBenchmark struct {
	LatencyMs     float64 `json:"latency_ms"`     // 时延，单位毫秒
	BandwidthMbps float64 `json:"bandwidth_mbps"` // 带宽，单位Mbps
}
