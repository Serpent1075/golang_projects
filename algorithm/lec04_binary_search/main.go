// 반드시 정렬 후 사용
package main

import "fmt"

func binarySearch(nums []int, target int) int {
	left, right := 0, len(nums)-1

	for left <= right {
		mid := left + (right-left)/2 // 탐색 범위의 중간 인덱스 계산

		if nums[mid] == target {
			return mid // target을 찾았으므로 해당 인덱스 반환
		} else if nums[mid] < target {
			left = mid + 1 // 탐색 범위를 오른쪽 절반으로 좁힘
		} else {
			right = mid - 1 // 탐색 범위를 왼쪽 절반으로 좁힘
		}
	}

	return -1 // target을 찾지 못했으므로 -1 반환
}

func binary_search(array []int64, to_search int64) int64 {
	found := -1
	low := 0
	high := len(array) - 1
	for low <= high {
		mid := (low + high) / 2
		if array[mid] == to_search {
			found = mid
			return int64(found)
		}
		if array[mid] > to_search {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return -1
}

func main() {
	nums := []int64{1, 3, 5, 7, 9}
	target := int64(5)

	index := binary_search(nums, target)
	if index != -1 {
		fmt.Printf("%d은(는) 배열의 %d번째 인덱스에 있습니다.\n", target, index)
	} else {
		fmt.Printf("%d은(는) 배열에 없습니다.\n", target)
	}
}
