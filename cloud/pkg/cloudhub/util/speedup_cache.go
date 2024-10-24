package util

import (
	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
	"sync"
)

// SpeedupCache 包含红黑树和 probeMap
type SpeedupCache struct {
	rbtree *redblacktree.Tree // 红黑树存储节点信息

	// CacheMap, key 是 Node.Name，value 是 probe
	probeMap sync.Map
}

// NewSpeedupCache 初始化 SpeedupCache
func NewSpeedupCache() *SpeedupCache {
	// 创建红黑树，比较器基于 Speedup 值（浮点数比较器）
	return &SpeedupCache{
		rbtree: redblacktree.NewWith(utils.Float64Comparator),
	}
}

func CalculateSpeedup(probe *Probe) float64 {
	// 假设 Speedup 是 CPU、内存和磁盘 IOPS 的加权和
	// 你可以根据具体的业务逻辑设计更复杂的计算公式
	return (probe.CPU.EventsPerSecond * 0.4) + (probe.Memory.OperationsPerSecond * 0.3) + (float64(probe.Disk.IOPS) * 0.3)
}

// AddProbe 向 probeMap 添加新节点，并根据 Probe 更新 rbtree 中的 Speedup
func (cache *SpeedupCache) AddProbe(nodeName string, probe Probe) {
	// 计算新的 Speedup
	probe.Speedup = CalculateSpeedup(&probe)

	// 添加到 probeMap
	cache.probeMap.Store(nodeName, probe)

	// 将节点添加到红黑树中
	cache.rbtree.Put(probe.Speedup, nodeName)
}

// AddRBTNode 向 rbtree 添加新节点，并根据 nodeName 更新 probeMap 中的 Probe
func (cache *SpeedupCache) AddRBTNode(nodeName string, speedup float64) {
	// 先从 probeMap 获取 Probe
	value, ok := cache.probeMap.Load(nodeName)
	if ok {
		probe := value.(Probe)
		// 更新 Probe 的 Speedup
		probe.Speedup = speedup

		// 更新红黑树
		cache.rbtree.Put(speedup, nodeName)

		// 更新 probeMap
		cache.probeMap.Store(nodeName, probe)
	}
}

// DeleteProbe 从 probeMap 删除节点，并从 rbtree 删除 Speedup 相关的节点
func (cache *SpeedupCache) DeleteProbe(nodeName string) {
	// 先从 probeMap 获取 Probe
	value, ok := cache.probeMap.Load(nodeName)
	if ok {
		probe := value.(Probe)

		// 从 probeMap 删除
		cache.probeMap.Delete(nodeName)

		// 从 rbtree 删除
		cache.rbtree.Remove(probe.Speedup)
	}
}

// DeleteRBTNode 从 rbtree 删除节点，并从 probeMap 中删除相关 Probe
func (cache *SpeedupCache) DeleteRBTNode(speedup float64) {
	// 先从 rbtree 获取 nodeName
	value, found := cache.rbtree.Get(speedup)
	if found {
		nodeName := value.(string)

		// 从 rbtree 删除
		cache.rbtree.Remove(speedup)

		// 从 probeMap 删除
		cache.probeMap.Delete(nodeName)
	}
}

// UpdateProbe 更新 probeMap 中的 Probe，并同步更新 rbtree 中的 Speedup
func (cache *SpeedupCache) UpdateProbe(nodeName string, newProbe Probe) {
	// 先从 probeMap 获取旧的 Probe
	oldValue, exists := cache.probeMap.Load(nodeName)
	if exists {
		oldProbe := oldValue.(Probe)
		// 如果存在旧的 Speedup 值，从 rbtree 中删除对应的旧条目
		cache.rbtree.Remove(oldProbe.Speedup)
	}

	// 计算新的 Speedup
	newProbe.Speedup = CalculateSpeedup(&newProbe)

	// 更新 probeMap
	cache.probeMap.Store(nodeName, newProbe)

	// 将新的节点添加到 rbtree
	cache.rbtree.Put(newProbe.Speedup, nodeName)
}

// UpdateRBTNode 更新 rbtree 中的 Speedup，并同步更新 probeMap 中的 Probe
func (cache *SpeedupCache) UpdateRBTNode(nodeName string, newSpeedup float64) {
	// 先从 probeMap 获取 Probe
	value, ok := cache.probeMap.Load(nodeName)
	if ok {
		probe := value.(Probe)

		// 更新 Probe 中的 Speedup
		probe.Speedup = newSpeedup

		// 更新红黑树
		cache.rbtree.Put(newSpeedup, nodeName)

		// 更新 probeMap
		cache.probeMap.Store(nodeName, probe)
	}
}

// GetProbe 从 probeMap 获取 Probe
func (cache *SpeedupCache) GetProbe(nodeName string) (Probe, bool) {
	value, ok := cache.probeMap.Load(nodeName)
	if ok {
		return value.(Probe), true
	}
	return Probe{}, false
}

// GetRBTNode 从 rbtree 获取节点
func (cache *SpeedupCache) GetRBTNode(speedup float64) (string, bool) {
	value, found := cache.rbtree.Get(speedup)
	if found {
		return value.(string), true
	}
	return "", false
}
