//최단 경로 찾기

package main

import "fmt"

type Point struct {
	row, col int
}

func bfs(maze [][]int) int {
	n, m := len(maze), len(maze[0])
	visited := make([][]bool, n)
	for i := range visited {
		visited[i] = make([]bool, m)
	}

	queue := []Point{{0, 0}}
	visited[0][0] = true
	distance := 0

	// 상, 하, 좌, 우 이동을 위한 델타 배열
	dr := []int{-1, 1, 0, 0}
	dc := []int{0, 0, -1, 1}

	for len(queue) > 0 {
		size := len(queue)
		for i := 0; i < size; i++ {
			curr := queue[0]
			queue = queue[1:]

			if curr.row == n-1 && curr.col == m-1 {
				return distance // 목표 위치에 도착하면 거리를 반환
			}

			for j := 0; j < 4; j++ {
				nextRow, nextCol := curr.row+dr[j], curr.col+dc[j]
				// 범위 체크 및 방문 여부, 벽 여부 확인
				if nextRow >= 0 && nextRow < n && nextCol >= 0 && nextCol < m &&
					!visited[nextRow][nextCol] && maze[nextRow][nextCol] == 0 {
					visited[nextRow][nextCol] = true
					queue = append(queue, Point{nextRow, nextCol})
				}
			}
		}
		distance++
	}

	return -1 // 목표 위치까지 이동할 수 없는 경우
}

func main() {
	maze := [][]int{
		{0, 0, 1, 0},
		{1, 0, 0, 0},
		{0, 0, 1, 0},
		{1, 0, 0, 0},
	}

	result := bfs(maze)
	fmt.Println("최단 경로 길이:", result)
}
