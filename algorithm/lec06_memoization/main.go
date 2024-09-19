package main

import "fmt"

// 메모이제이션을 위한 맵
var memo = map[uint]uint{}
var MOD = 10007

func fibonacci(n uint) uint {
	if n == 0 {
		return 0
	} else if n == 1 {
		return 1
	} else if val, ok := memo[n]; ok {
		return val // 이미 계산된 값이 있으면 재활용
	} else {
		result := fibonacci(n-1) + fibonacci(n-2)
		memo[n] = result // 계산된 값을 저장
		return result
	}
}

func main() {
	var n uint = 200
	result := fibonacci(n)
	fmt.Printf("%d번째 피보나치 수: %d\n", n, result)
}
