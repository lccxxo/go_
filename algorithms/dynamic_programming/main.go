package main

import "fmt"

// 动态规划算法示例

// 1. 斐波那契数列 - 动态规划版本
// 时间复杂度: O(n), 空间复杂度: O(n)
func FibonacciDP(n int) int {
	if n <= 1 {
		return n
	}

	dp := make([]int, n+1)
	dp[0] = 0
	dp[1] = 1

	for i := 2; i <= n; i++ {
		dp[i] = dp[i-1] + dp[i-2]
	}
	return dp[n]
}

// 斐波那契数列 - 空间优化版本 O(1)
func FibonacciOptimized(n int) int {
	if n <= 1 {
		return n
	}

	prev, curr := 0, 1
	for i := 2; i <= n; i++ {
		prev, curr = curr, prev+curr
	}
	return curr
}

// 2. 爬楼梯问题：每次可以爬1或2个台阶，有多少种方法爬到n层
func ClimbStairs(n int) int {
	if n <= 2 {
		return n
	}

	dp := make([]int, n+1)
	dp[1] = 1
	dp[2] = 2

	for i := 3; i <= n; i++ {
		dp[i] = dp[i-1] + dp[i-2]
	}
	return dp[n]
}

// 3. 最长公共子序列 (LCS)
func LongestCommonSubsequence(text1, text2 string) int {
	m, n := len(text1), len(text2)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if text1[i-1] == text2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = max(dp[i-1][j], dp[i][j-1])
			}
		}
	}
	return dp[m][n]
}

// 4. 0-1背包问题
func Knapsack01(weights []int, values []int, capacity int) int {
	n := len(weights)
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, capacity+1)
	}

	for i := 1; i <= n; i++ {
		for w := 1; w <= capacity; w++ {
			if weights[i-1] <= w {
				dp[i][w] = max(dp[i-1][w], dp[i-1][w-weights[i-1]]+values[i-1])
			} else {
				dp[i][w] = dp[i-1][w]
			}
		}
	}
	return dp[n][capacity]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	// 斐波那契数列
	fmt.Println("=== 斐波那契数列 ===")
	n := 10
	fmt.Printf("FibonacciDP(%d) = %d\n", n, FibonacciDP(n))
	fmt.Printf("FibonacciOptimized(%d) = %d\n", n, FibonacciOptimized(n))
	fmt.Println()

	// 爬楼梯
	fmt.Println("=== 爬楼梯问题 ===")
	stairs := 5
	fmt.Printf("爬 %d 层楼梯的方法数: %d\n", stairs, ClimbStairs(stairs))
	fmt.Println()

	// 最长公共子序列
	fmt.Println("=== 最长公共子序列 ===")
	text1, text2 := "abcde", "ace"
	lcs := LongestCommonSubsequence(text1, text2)
	fmt.Printf("'%s' 和 '%s' 的最长公共子序列长度: %d\n", text1, text2, lcs)
	fmt.Println()

	// 0-1背包
	fmt.Println("=== 0-1背包问题 ===")
	weights := []int{2, 3, 4, 5}
	values := []int{3, 4, 5, 6}
	capacity := 8
	maxValue := Knapsack01(weights, values, capacity)
	fmt.Printf("物品重量: %v\n", weights)
	fmt.Printf("物品价值: %v\n", values)
	fmt.Printf("背包容量: %d\n", capacity)
	fmt.Printf("最大价值: %d\n", maxValue)
}
