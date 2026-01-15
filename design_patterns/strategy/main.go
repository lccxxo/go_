package main

import "fmt"

// 策略模式：定义一系列算法，把它们封装起来，并且使它们可以互相替换

// 排序策略接口
type SortStrategy interface {
	Sort(data []int) []int
}

// 冒泡排序策略
type BubbleSort struct{}

func (b *BubbleSort) Sort(data []int) []int {
	result := make([]int, len(data))
	copy(result, data)
	n := len(result)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if result[j] > result[j+1] {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}
	return result
}

// 快速排序策略
type QuickSort struct{}

func (q *QuickSort) Sort(data []int) []int {
	result := make([]int, len(data))
	copy(result, data)
	q.quickSort(result, 0, len(result)-1)
	return result
}

func (q *QuickSort) quickSort(data []int, left, right int) {
	if left < right {
		pivot := q.partition(data, left, right)
		q.quickSort(data, left, pivot-1)
		q.quickSort(data, pivot+1, right)
	}
}

func (q *QuickSort) partition(data []int, left, right int) int {
	pivot := data[right]
	i := left - 1
	for j := left; j < right; j++ {
		if data[j] < pivot {
			i++
			data[i], data[j] = data[j], data[i]
		}
	}
	data[i+1], data[right] = data[right], data[i+1]
	return i + 1
}

// 排序上下文
type SortContext struct {
	strategy SortStrategy
}

func NewSortContext(strategy SortStrategy) *SortContext {
	return &SortContext{strategy: strategy}
}

func (sc *SortContext) SetStrategy(strategy SortStrategy) {
	sc.strategy = strategy
}

func (sc *SortContext) ExecuteSort(data []int) []int {
	return sc.strategy.Sort(data)
}

func main() {
	data := []int{64, 34, 25, 12, 22, 11, 90}
	fmt.Println("原始数据:", data)

	// 使用冒泡排序
	bubbleSort := &BubbleSort{}
	context := NewSortContext(bubbleSort)
	result1 := context.ExecuteSort(data)
	fmt.Println("冒泡排序结果:", result1)

	// 切换到快速排序
	quickSort := &QuickSort{}
	context.SetStrategy(quickSort)
	result2 := context.ExecuteSort(data)
	fmt.Println("快速排序结果:", result2)
}
