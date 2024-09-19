package main

import "fmt"

func main() {
    // 5개의 정점을 가진 그래프
    numVertices := 5

    // 인접 행렬 초기화 (모든 간선 가중치를 0으로 초기화)
    adjacencyMatrix := make([][]int, numVertices)
    for i := 0; i < numVertices; i++ {
        adjacencyMatrix[i] = make([]int, numVertices)
    }

    // 간선 추가 (무방향 그래프)
    adjacencyMatrix[0][1] = 1
    adjacencyMatrix[1][0] = 1
    adjacencyMatrix[1][2] = 1
    adjacencyMatrix[2][1] = 1
    adjacencyMatrix[2][3] = 1
    adjacencyMatrix[3][2] = 1
    adjacencyMatrix[3][4] = 1
    adjacencyMatrix[4][3] = 1

    // 인접 행렬 출력
    fmt.Println("인접 행렬:")
    for _, row := range adjacencyMatrix {
        fmt.Println(row)
    }
}