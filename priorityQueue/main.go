package main

import (
	"container/heap"
	"fmt"
)

// 模拟请求
type Request struct {
	ID       int    // 唯一标识ID
	Content  string // 请求内容
	Priority int    // 请求优先级，越小优先级越高。
}

// 优先级队列
type PriorityQueue []*Request

// 获取队列数
func (p PriorityQueue) Len() int {
	return len(p)
}

// 比较优先级
func (p PriorityQueue) Less(i, j int) bool {
	return p[i].Priority < p[j].Priority
}

// 进行位置交换
func (p PriorityQueue) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

// Push 添加请求到队列
func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*Request))
}

// Pop 移除并返回优先级最高的请求
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	request := old[n-1]
	*pq = old[0 : n-1]
	return request
}

func main() {
	p := &PriorityQueue{}
	heap.Init(p)

	heap.Push(p, &Request{ID: 1, Content: "普通请求", Priority: 3})
	heap.Push(p, &Request{ID: 2, Content: "VIP请求", Priority: 1})
	heap.Push(p, &Request{ID: 3, Content: "普通请求", Priority: 2})

	fmt.Println("处理请求：")
	for p.Len() > 0 {
		req := heap.Pop(p).(*Request)
		fmt.Printf("处理请求 ID: %d, 内容: %s, 优先级: %d\n", req.ID, req.Content, req.Priority)
	}
}
