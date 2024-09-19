package main

import "fmt"

func f(nums []int, x int) bool {
	var sum int
	for _, num := range nums {
		sum += num % x
	}
	return sum <= 3 // 조건: 나머지의 합이 3 이하
}

func parametricSearch(nums []int, left, right int) int {
	for left < right {
		mid := left + (right-left+1)/2 // mid를 올림하여 무한 루프 방지
		if f(nums, mid) {
			left = mid // mid가 조건을 만족하면 탐색 범위를 오른쪽으로 이동
		} else {
			right = mid - 1 // mid가 조건을 만족하지 않으면 탐색 범위를 왼쪽으로 이동
		}
	}
	return left // left는 f(x)가 true인 가장 큰 x 값
}

func main() {
	nums := []int{1, 3, 5, 7, 9}
	left, right := 1, 9 // 탐색 범위 설정 (1부터 9까지)
	result := parametricSearch(nums, left, right)
	fmt.Println(result)
}

func minMax(arr []int) (r1 int, r2 int) {
	min := arr[0]
	max := arr[0]

	for value := range arr {
		if min > arr[value] {
			min = arr[value]
		}

		if max < arr[value] {
			max = arr[value]
		}
	}

	return min, max
}
