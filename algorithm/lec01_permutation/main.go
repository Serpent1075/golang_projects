package main

import (
	"fmt"
)

func permute(nums []int, k, limit int) [][]int {
	var result [][]int
	var backtrack func(start, count, currentSum int, currentCombination []int)

	backtrack = func(start, count, currentSum int, currentCombination []int) {
		// 현재까지 선택한 숫자 개수가 k개이고 합이 limit 이하인 경우 결과에 추가
		if count == k && currentSum <= limit {
			result = append(result, append([]int(nil), currentCombination...))
			return
		}
		// 현재까지 선택한 숫자 개수가 k개를 초과하거나 합이 limit을 초과하는 경우 가지치기
		if count > k || currentSum > limit {
			return
		}
		// start 위치부터 끝까지 반복하며 조합 생성
		for i := start; i < len(nums); i++ {
			backtrack(i+1, count+1, currentSum+nums[i], append(currentCombination, nums[i]))
		}
	}

	backtrack(0, 0, 0, []int{})
	return result
}

func main() {
	nums := []int{20, 7, 23, 19, 10, 15, 25, 8, 13}
	k := 7       // 선택할 숫자 개수
	limit := 100 // 합의 상한
	result := permute(nums, k, limit)
	for _, v := range result {
		//sort.Ints(v)
		//bubbleSort(v)
		fmt.Println(v)
	}

}

func bubbleSort(arr []int) {
	n := len(arr)
	swapped := true

	for swapped {
		swapped = false
		for i := 0; i < n-1; i++ {
			if arr[i] > arr[i+1] {
				// swap arr[i] and arr[i+1]
				arr[i], arr[i+1] = arr[i+1], arr[i]
				swapped = true
			}
		}
		n--
	}
}
