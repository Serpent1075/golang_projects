package main

import "fmt"

func main() {
    // 5개의 정점을 가진 그래프
    numVertices := 5

    // 인접 리스트 초기화
    adjacencyList := make([][]int, numVertices)
    for i := 0; i < numVertices; i++ {
        adjacencyList[i] = make([]int, 0)
    }

    // 간선 추가 (무방향 그래프)
    adjacencyList[0] = append(adjacencyList[0], 1)
    adjacencyList[1] = append(adjacencyList[1], 0, 2)
    adjacencyList[2] = append(adjacencyList[2], 1, 3)
    adjacencyList[3] = append(adjacencyList[3], 2, 4)
    adjacencyList[4] = append(adjacencyList[4], 3)

    // 인접 리스트 출력
    fmt.Println("인접 리스트:")
    for i, neighbors := range adjacencyList {
        fmt.Printf("%d: %v\n", i, neighbors)
    }
}