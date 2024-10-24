package main

import (
	"fmt"
	rbt "github.com/emirpasic/gods/trees/redblacktree"
)

// 节点结构体，包含节点名称和加速比
type NodeData struct {
	NodeName string
	Speedup  float64
}

// 插入或更新节点的加速比
func updateSpeedup(tree *rbt.Tree, nodeName string, newSpeedup float64) {
	// 首先检查节点是否已经存在
	if value, found := tree.Get(nodeName); found {
		// 如果找到节点，更新其加速比
		nodeData := value.(*NodeData)
		nodeData.Speedup = newSpeedup
		fmt.Printf("更新节点 %s 的加速比为 %.2f\n", nodeData.NodeName, nodeData.Speedup)
	} else {
		// 如果节点不存在，插入新节点
		newNode := &NodeData{NodeName: nodeName, Speedup: newSpeedup}
		tree.Put(nodeName, newNode)
		fmt.Printf("插入节点 %s, 加速比为 %.2f\n", newNode.NodeName, newNode.Speedup)
	}
}
