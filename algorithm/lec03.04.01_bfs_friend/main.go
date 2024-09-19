package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func shortestPath(graph [][]int, start, end int) int {
	n := len(graph)
	visited := make([]bool, n)
	queue := []int{start}
	visited[start] = true
	distance := 0

	for len(queue) > 0 {
		size := len(queue)
		for i := 0; i < size; i++ {
			person := queue[0]
			queue = queue[1:]

			if person == end {
				return distance
			}

			for _, friend := range graph[person] {
				if !visited[friend] {
					visited[friend] = true
					queue = append(queue, friend)
				}
			}
		}
		distance++
	}

	return -1 // 연결되지 않은 경우
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("사람 수 N 입력: ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	var n int
	fmt.Sscan(text, &n)

	graph := make([][]int, n)
	for i := 0; i < n; i++ {
		fmt.Printf("%d번 사람의 친구 목록 입력: ", i)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		friends := strings.Split(text, " ")
		for _, friendStr := range friends {
			friend, _ := strconv.Atoi(friendStr)
			graph[i] = append(graph[i], friend)
		}
	}

	fmt.Print("시작 사람 A, 도착 사람 B 입력: ")
	text, _ = reader.ReadString('\n')
	text = strings.TrimSpace(text)
	var a, b int
	fmt.Sscan(text, &a, &b)

	result := shortestPath(graph, a, b)
	if result == -1 {
		fmt.Println("연결되지 않았습니다.")
	} else {
		fmt.Printf("최단 거리: %d\n", result)
	}
}