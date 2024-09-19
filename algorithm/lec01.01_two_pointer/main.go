//특징	            인접 행렬				 인접 리스트
//공간 복잡도		O(V^2) (V: 정점의 개수)	  O(V + E) (V: 정점의 개수, E: 간선의 개수)
//특정 간선 확인	O(1) 시간에 가능		  O(degree(u)) 시간에 가능 (degree(u): 정점 u에 연결된 간선의 개수)
//정점에 연결된 모든 간선 탐색	O(V) 시간에 가능	O(degree(u)) 시간에 가능
//새로운 간선 추가/삭제	O(1) 시간에 가능	O(degree(u)) 또는 O(1) 시간에 가능 (구현 방식에 따라 다름)
//적합한 경우	밀집 그래프 (간선의 수가 많을 때)	희소 그래프 (간선의 수가 적을 때)
//장점	간선 존재 여부 및 가중치 확인이 빠름, 구현이 간단함	메모리 사용량이 적음, 특정 정점에 연결된 간선 탐색이 빠름
//단점	메모리 사용량이 많음, 희소 그래프에서 메모리 낭비 발생, 간선 탐색이 느림	구현이 복잡할 수 있음, 특정 간선 존재 여부 확인이 느림, 새로운 간선 추가/삭제가 느릴 수 있음

package main

import "fmt"

func findSubarrayWithSum(nums []int, target int) []int {
	left, right := 0, 0
	currentSum := 0

	for right < len(nums) {
		currentSum += nums[right]

		for currentSum > target && left < right {
			currentSum -= nums[left]
			left++
		}

		if currentSum == target {
			return nums[left : right+1]
		}

		right++
	}

	return nil // 합이 target인 부분 배열을 찾지 못한 경우
}

func main() {
	nums := []int{1, 4, 20, 3, 10, 5}
	target := 33
	result := findSubarrayWithSum(nums, target)

	if result != nil {
		fmt.Println("합이", target, "인 부분 배열:", result)
	} else {
		fmt.Println("합이", target, "인 부분 배열을 찾지 못했습니다.")
	}
}
