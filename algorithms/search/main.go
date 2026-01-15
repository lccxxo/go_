package main

import "fmt"

// 查找算法示例

// 线性查找 O(n)
func LinearSearch(arr []int, target int) int {
	for i, v := range arr {
		if v == target {
			return i
		}
	}
	return -1
}

// 二分查找 O(log n) - 要求数组已排序
func BinarySearch(arr []int, target int) int {
	left, right := 0, len(arr)-1

	for left <= right {
		mid := left + (right-left)/2

		if arr[mid] == target {
			return mid
		} else if arr[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return -1
}

// 递归版本的二分查找
func BinarySearchRecursive(arr []int, target, left, right int) int {
	if left > right {
		return -1
	}

	mid := left + (right-left)/2

	if arr[mid] == target {
		return mid
	} else if arr[mid] < target {
		return BinarySearchRecursive(arr, target, mid+1, right)
	} else {
		return BinarySearchRecursive(arr, target, left, mid-1)
	}
}

func main() {
	// 线性查找
	arr1 := []int{3, 7, 1, 9, 2, 5, 8}
	target1 := 5
	index1 := LinearSearch(arr1, target1)
	fmt.Printf("线性查找: 在数组 %v 中查找 %d，结果: ", arr1, target1)
	if index1 != -1 {
		fmt.Printf("找到，索引为 %d\n", index1)
	} else {
		fmt.Println("未找到")
	}

	// 二分查找（需要有序数组）
	arr2 := []int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19}
	target2 := 7
	index2 := BinarySearch(arr2, target2)
	fmt.Printf("二分查找: 在数组 %v 中查找 %d，结果: ", arr2, target2)
	if index2 != -1 {
		fmt.Printf("找到，索引为 %d\n", index2)
	} else {
		fmt.Println("未找到")
	}

	// 递归二分查找
	target3 := 15
	index3 := BinarySearchRecursive(arr2, target3, 0, len(arr2)-1)
	fmt.Printf("递归二分查找: 在数组 %v 中查找 %d，结果: ", arr2, target3)
	if index3 != -1 {
		fmt.Printf("找到，索引为 %d\n", index3)
	} else {
		fmt.Println("未找到")
	}
}
