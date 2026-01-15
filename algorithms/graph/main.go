package main

import "fmt"

// 图算法示例

// 图的邻接表表示
type Graph struct {
	Vertices int
	AdjList  map[int][]int
}

func NewGraph(vertices int) *Graph {
	return &Graph{
		Vertices: vertices,
		AdjList:  make(map[int][]int),
	}
}

func (g *Graph) AddEdge(u, v int) {
	g.AdjList[u] = append(g.AdjList[u], v)
	// 如果是无向图，取消下面的注释
	// g.AdjList[v] = append(g.AdjList[v], u)
}

// 深度优先搜索 (DFS)
func (g *Graph) DFS(start int) []int {
	visited := make(map[int]bool)
	result := make([]int, 0)

	var dfsHelper func(int)
	dfsHelper = func(v int) {
		visited[v] = true
		result = append(result, v)

		for _, neighbor := range g.AdjList[v] {
			if !visited[neighbor] {
				dfsHelper(neighbor)
			}
		}
	}

	dfsHelper(start)
	return result
}

// 广度优先搜索 (BFS)
func (g *Graph) BFS(start int) []int {
	visited := make(map[int]bool)
	queue := []int{start}
	result := make([]int, 0)
	visited[start] = true

	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		result = append(result, v)

		for _, neighbor := range g.AdjList[v] {
			if !visited[neighbor] {
				visited[neighbor] = true
				queue = append(queue, neighbor)
			}
		}
	}
	return result
}

// 拓扑排序（用于有向无环图）
func (g *Graph) TopologicalSort() []int {
	inDegree := make(map[int]int)
	for i := 0; i < g.Vertices; i++ {
		inDegree[i] = 0
	}

	// 计算入度
	for _, neighbors := range g.AdjList {
		for _, v := range neighbors {
			inDegree[v]++
		}
	}

	// 找到所有入度为0的节点
	queue := make([]int, 0)
	for i := 0; i < g.Vertices; i++ {
		if inDegree[i] == 0 {
			queue = append(queue, i)
		}
	}

	result := make([]int, 0)
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		result = append(result, u)

		for _, v := range g.AdjList[u] {
			inDegree[v]--
			if inDegree[v] == 0 {
				queue = append(queue, v)
			}
		}
	}

	return result
}

// 最短路径：Dijkstra算法（单源最短路径）
func (g *Graph) Dijkstra(start int, weights map[string]int) map[int]int {
	dist := make(map[int]int)
	visited := make(map[int]bool)

	// 初始化距离
	for i := 0; i < g.Vertices; i++ {
		dist[i] = 999999 // 无穷大
	}
	dist[start] = 0

	for i := 0; i < g.Vertices; i++ {
		// 找到未访问的距离最小的节点
		u := -1
		minDist := 999999
		for v := 0; v < g.Vertices; v++ {
			if !visited[v] && dist[v] < minDist {
				minDist = dist[v]
				u = v
			}
		}

		if u == -1 {
			break
		}

		visited[u] = true

		// 更新邻居节点的距离
		for _, v := range g.AdjList[u] {
			key := fmt.Sprintf("%d-%d", u, v)
			weight := weights[key]
			if weight == 0 {
				weight = 1 // 默认权重为1
			}

			if !visited[v] && dist[u]+weight < dist[v] {
				dist[v] = dist[u] + weight
			}
		}
	}

	return dist
}

func main() {
	// 创建图
	g := NewGraph(5)
	g.AddEdge(0, 1)
	g.AddEdge(0, 2)
	g.AddEdge(1, 3)
	g.AddEdge(2, 4)
	g.AddEdge(3, 4)

	fmt.Println("=== 深度优先搜索 (DFS) ===")
	dfsResult := g.DFS(0)
	fmt.Printf("从节点 0 开始的DFS: %v\n", dfsResult)
	fmt.Println()

	fmt.Println("=== 广度优先搜索 (BFS) ===")
	bfsResult := g.BFS(0)
	fmt.Printf("从节点 0 开始的BFS: %v\n", bfsResult)
	fmt.Println()

	// 拓扑排序示例
	fmt.Println("=== 拓扑排序 ===")
	g2 := NewGraph(6)
	g2.AddEdge(5, 2)
	g2.AddEdge(5, 0)
	g2.AddEdge(4, 0)
	g2.AddEdge(4, 1)
	g2.AddEdge(2, 3)
	g2.AddEdge(3, 1)
	topoResult := g2.TopologicalSort()
	fmt.Printf("拓扑排序结果: %v\n", topoResult)
	fmt.Println()

	// Dijkstra算法示例
	fmt.Println("=== Dijkstra最短路径 ===")
	g3 := NewGraph(5)
	g3.AddEdge(0, 1)
	g3.AddEdge(0, 2)
	g3.AddEdge(1, 3)
	g3.AddEdge(2, 4)
	g3.AddEdge(3, 4)
	weights := map[string]int{
		"0-1": 4,
		"0-2": 2,
		"1-3": 5,
		"2-4": 3,
		"3-4": 1,
	}
	distances := g3.Dijkstra(0, weights)
	fmt.Printf("从节点 0 到各节点的最短距离: %v\n", distances)
}
