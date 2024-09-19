package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("N M K 입력: ")
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	var n, m, k int
	fmt.Sscan(text, &n, &m, &k)

	board := make([][]byte, n)
	for i := 0; i < n; i++ {
		board[i], _ = reader.ReadBytes('\n')
		board[i] = board[i][:len(board[i])-1] // 개행 문자 제거
	}

	// 누적합 배열 생성 (B와 W 각각에 대해)
	prefixSumB := make([][]int, n+1)
	prefixSumW := make([][]int, n+1)
	for i := 0; i <= n; i++ {
		prefixSumB[i] = make([]int, m+1)
		prefixSumW[i] = make([]int, m+1)
	}

	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if board[i-1][j-1] == 'B' {
				prefixSumB[i][j] = 1
			} else {
				prefixSumW[i][j] = 1
			}
			prefixSumB[i][j] += prefixSumB[i-1][j] + prefixSumB[i][j-1] - prefixSumB[i-1][j-1]
			prefixSumW[i][j] += prefixSumW[i-1][j] + prefixSumW[i][j-1] - prefixSumW[i-1][j-1]
		}
	}

	// K x K 영역의 B와 W 개수를 누적합 배열을 이용하여 계산하고 최솟값 찾기
	minCount := k * k
	for i := 0; i <= n-k; i++ {
		for j := 0; j <= m-k; j++ {
			countB := prefixSumB[i+k][j+k] - prefixSumB[i+k][j] - prefixSumB[i][j+k] + prefixSumB[i][j]
			countW := prefixSumW[i+k][j+k] - prefixSumW[i+k][j] - prefixSumW[i][j+k] + prefixSumW[i][j]
			minCount = min(minCount, min(countB, countW))
		}
	}

	fmt.Println(minCount)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}