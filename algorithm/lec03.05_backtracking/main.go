package main

import "fmt"

func subsets(nums []int) [][]int {
	var result [][]int
	var backtrack func(start int, currentSubset []int)

	backtrack = func(start int, currentSubset []int) {
		// 현재 부분집합을 결과에 추가
		result = append(result, append([]int(nil), currentSubset...))

		// start 위치부터 끝까지 반복하며 부분집합 생성
		for i := start; i < len(nums); i++ {
			backtrack(i+1, append(currentSubset, nums[i]))
		}
	}

	backtrack(0, []int{})
	return result
}

func main() {
	nums := []int{1, 2, 3}
	result := subsets(nums)
	fmt.Println(result)
}
