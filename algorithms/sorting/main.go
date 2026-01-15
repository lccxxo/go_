package main

import "fmt"

// 排序算法示例

// 冒泡排序 O(n²)
func BubbleSort(arr []int) []int {
	result := make([]int, len(arr))
	copy(result, arr)
	n := len(result)

	for i := 0; i < n-1; i++ {
		swapped := false
		for j := 0; j < n-i-1; j++ {
			if result[j] > result[j+1] {
				result[j], result[j+1] = result[j+1], result[j]
				swapped = true
			}
		}
		if !swapped {
			break // 优化：如果这一轮没有交换，说明已经有序
		}
	}
	return result
}

// 快速排序 O(n log n) 平均，O(n²) 最坏
func QuickSort(arr []int) []int {
	result := make([]int, len(arr))
	copy(result, arr)
	quickSort(result, 0, len(result)-1)
	return result
}

func quickSort(arr []int, left, right int) {
	if left < right {
		pivot := partition(arr, left, right)
		quickSort(arr, left, pivot-1)
		quickSort(arr, pivot+1, right)
	}
}

func partition(arr []int, left, right int) int {
	pivot := arr[right]
	i := left - 1

	for j := left; j < right; j++ {
		if arr[j] < pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[right] = arr[right], arr[i+1]
	return i + 1
}

// 归并排序 O(n log n)
func MergeSort(arr []int) []int {
	if len(arr) <= 1 {
		return arr
	}

	mid := len(arr) / 2
	left := MergeSort(arr[:mid])
	right := MergeSort(arr[mid:])

	return merge(left, right)
}

func merge(left, right []int) []int {
	result := make([]int, 0, len(left)+len(right))
	i, j := 0, 0

	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}

	result = append(result, left[i:]...)
	result = append(result, right[j:]...)
	return result
}

// 插入排序 O(n²)
func InsertionSort(arr []int) []int {
	result := make([]int, len(arr))
	copy(result, arr)

	for i := 1; i < len(result); i++ {
		key := result[i]
		j := i - 1

		for j >= 0 && result[j] > key {
			result[j+1] = result[j]
			j--
		}
		result[j+1] = key
	}
	return result
}

// 选择排序 O(n²)
func SelectionSort(arr []int) []int {
	result := make([]int, len(arr))
	copy(result, arr)

	for i := 0; i < len(result)-1; i++ {
		minIdx := i
		for j := i + 1; j < len(result); j++ {
			if result[j] < result[minIdx] {
				minIdx = j
			}
		}
		result[i], result[minIdx] = result[minIdx], result[i]
	}
	return result
}

func main() {
	arr := []int{64, 34, 25, 12, 22, 11, 90, 5}
	fmt.Println("原始数组:", arr)
	fmt.Println()

	fmt.Println("冒泡排序:", BubbleSort(arr))
	fmt.Println("快速排序:", QuickSort(arr))
	fmt.Println("归并排序:", MergeSort(arr))
	fmt.Println("插入排序:", InsertionSort(arr))
	fmt.Println("选择排序:", SelectionSort(arr))
}
