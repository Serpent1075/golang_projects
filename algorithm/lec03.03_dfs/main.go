//깊이 우선 탐색

package main

import "fmt"

func dfsStack(graph [][]int, start int) {
	numVertices := len(graph)
	visited := make([]bool, numVertices)
	stack := []int{start}

	for len(stack) > 0 {
		// 스택에서 정점 추출
		vertex := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if !visited[vertex] {
			visited[vertex] = true
			fmt.Print(vertex, " ")

			// 인접 정점을 스택에 추가 (역순으로 추가하여 낮은 번호의 정점부터 방문)
			for i := len(graph[vertex]) - 1; i >= 0; i-- {
				neighbor := graph[vertex][i]
				if !visited[neighbor] {
					stack = append(stack, neighbor)
				}
			}
		}
	}
}

func dfsRecursive(graph [][]int, vertex int, visited []bool) {
	visited[vertex] = true
	fmt.Print(vertex, " ")

	for _, neighbor := range graph[vertex] {
		if !visited[neighbor] {
			dfsRecursive(graph, neighbor, visited)
		}
	}
}

func main() {
	graph := [][]int{
		{1, 2},
		{0, 3, 4},
		{0, 4},
		{1},
		{1, 2},
	}

	fmt.Println("스택")
	startVertex := 0
	dfsStack(graph, startVertex) // 출력: 0 1 3 4 2

	////
	fmt.Println("\n재귀")
	numVertices := len(graph)
	visited := make([]bool, numVertices)

	startVertex = 0
	dfsRecursive(graph, startVertex, visited)
}
