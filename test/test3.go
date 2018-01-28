package main

import "fmt"

func dfsutil(graph [][]int, vertex int, visited []int) {
	visited[vertex] = 1
	for i := 0; i < len(graph[vertex]); i++ {
		if visited[graph[vertex][i]] == 0 {
			dfsutil(graph, graph[vertex][i], visited)
		}
	}
	return
}

func DFS(graph [][]int, comp *int) {
	var visited []int = make([]int, len(graph), len(graph))
	for i := 1; i < len(visited); i++ {
		if visited[i] == 0 {
			dfsutil(graph, i, visited)
			*comp = *comp + 1
		}
	}
	return
}

func main() {
	var n, m int
	fmt.Scanf("%d %d", &n, &m)
	var graph [][]int = make([][]int, n+1, n+1)
	var vertex1, vertex2 int
	for i := 0; i < m; i++ {
		fmt.Scanf("%d %d", &vertex1, &vertex2)
		graph[vertex1] = append(graph[vertex1], vertex2)
		graph[vertex2] = append(graph[vertex2], vertex1)
	}
	var comp int = 0
	DFS(graph, &comp)
	fmt.Printf("Number of connected components : %d\n", comp)
}
