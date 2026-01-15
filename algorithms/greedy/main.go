package main

import (
	"fmt"
	"sort"
)

// 贪心算法示例

// 1. 活动选择问题：选择最多的不重叠活动
type Activity struct {
	Start int
	End   int
}

func ActivitySelection(activities []Activity) []Activity {
	if len(activities) == 0 {
		return []Activity{}
	}

	// 按结束时间排序
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].End < activities[j].End
	})

	result := []Activity{activities[0]}
	lastEnd := activities[0].End

	for i := 1; i < len(activities); i++ {
		if activities[i].Start >= lastEnd {
			result = append(result, activities[i])
			lastEnd = activities[i].End
		}
	}

	return result
}

// 2. 找零问题：用最少的硬币找零
func CoinChange(coins []int, amount int) []int {
	// 按面额从大到小排序
	sort.Sort(sort.Reverse(sort.IntSlice(coins)))

	result := make([]int, 0)
	remaining := amount

	for _, coin := range coins {
		for remaining >= coin {
			result = append(result, coin)
			remaining -= coin
		}
		if remaining == 0 {
			break
		}
	}

	if remaining > 0 {
		return nil // 无法找零
	}
	return result
}

// 3. 分数背包问题（可以取部分物品）
type Item struct {
	Weight int
	Value  int
}

func FractionalKnapsack(items []Item, capacity int) float64 {
	// 按价值/重量比从大到小排序
	sort.Slice(items, func(i, j int) bool {
		ratioI := float64(items[i].Value) / float64(items[i].Weight)
		ratioJ := float64(items[j].Value) / float64(items[j].Weight)
		return ratioI > ratioJ
	})

	totalValue := 0.0
	remaining := capacity

	for _, item := range items {
		if remaining >= item.Weight {
			// 可以完全装入
			totalValue += float64(item.Value)
			remaining -= item.Weight
		} else {
			// 只能装入部分
			fraction := float64(remaining) / float64(item.Weight)
			totalValue += float64(item.Value) * fraction
			break
		}
	}

	return totalValue
}

func main() {
	// 活动选择问题
	fmt.Println("=== 活动选择问题 ===")
	activities := []Activity{
		{1, 4},
		{3, 5},
		{0, 6},
		{5, 7},
		{8, 9},
		{5, 9},
	}
	selected := ActivitySelection(activities)
	fmt.Printf("所有活动: %v\n", activities)
	fmt.Printf("选择的活动: %v\n", selected)
	fmt.Printf("最多可以选择 %d 个活动\n", len(selected))
	fmt.Println()

	// 找零问题
	fmt.Println("=== 找零问题 ===")
	coins := []int{1, 5, 10, 25}
	amount := 67
	change := CoinChange(coins, amount)
	if change != nil {
		fmt.Printf("找零 %d 元，使用硬币: %v\n", amount, change)
	} else {
		fmt.Printf("无法找零 %d 元\n", amount)
	}
	fmt.Println()

	// 分数背包问题
	fmt.Println("=== 分数背包问题 ===")
	items := []Item{
		{10, 60},  // 价值/重量 = 6
		{20, 100}, // 价值/重量 = 5
		{30, 120}, // 价值/重量 = 4
	}
	capacity := 50
	maxValue := FractionalKnapsack(items, capacity)
	fmt.Printf("物品: %v\n", items)
	fmt.Printf("背包容量: %d\n", capacity)
	fmt.Printf("最大价值: %.2f\n", maxValue)
}
