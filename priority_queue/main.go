package main

import (
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
func (pq PriorityQueue) Len() int {
	return len(pq)
}

// Push 添加请求到队列，按优先级插入到正确位置（优先级越小越靠前）
func (pq *PriorityQueue) Push(req *Request) {
	// 如果队列为空，直接添加
	if len(*pq) == 0 {
		*pq = append(*pq, req)
		return
	}

	// 找到插入位置：找到第一个优先级大于等于新请求的位置
	insertIndex := len(*pq)
	for i, r := range *pq {
		if r.Priority >= req.Priority {
			insertIndex = i
			break
		}
	}

	// 在指定位置插入
	*pq = append(*pq, nil)                           // 扩展切片
	copy((*pq)[insertIndex+1:], (*pq)[insertIndex:]) // 后移元素
	(*pq)[insertIndex] = req                         // 插入新元素
}

// Pop 移除并返回优先级最高的请求（队列第一个元素）
func (pq *PriorityQueue) Pop() *Request {
	if len(*pq) == 0 {
		return nil
	}

	// 获取第一个元素（优先级最高的）
	req := (*pq)[0]
	// 移除第一个元素
	*pq = (*pq)[1:]
	return req
}

func main() {
	p := &PriorityQueue{}

	// 添加请求，会自动按优先级排序
	p.Push(&Request{ID: 1, Content: "普通请求", Priority: 3})
	p.Push(&Request{ID: 2, Content: "SVIP请求", Priority: 1})
	p.Push(&Request{ID: 3, Content: "VIP请求", Priority: 2})

	fmt.Println("处理请求：")
	for p.Len() > 0 {
		// Pop() 会返回优先级最高的请求（Priority 值最小的）
		req := p.Pop()
		fmt.Printf("处理请求 ID: %d, 内容: %s, 优先级: %d\n", req.ID, req.Content, req.Priority)
	}
}
