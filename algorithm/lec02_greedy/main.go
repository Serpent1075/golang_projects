package main

import "fmt"

func minCoinChange(coins []int, amount int) map[int]int {
	result := make(map[int]int)
	// 큰 동전부터 탐욕적으로 선택
	for i := len(coins) - 1; i >= 0; i-- {
		/*
			for amount >= coins[i] {
				result[coins[i]]++
				amount -= coins[i]
			}
		*/
		result[coins[i]] = amount / coins[i]
		amount %= coins[i]
	}
	return result
}

func main() {
	coins := []int{1, 5, 10, 50, 100, 500}
	amount := 3620
	result := minCoinChange(coins, amount)

	for coin, count := range result {
		if count > 0 {
			fmt.Printf("%d원: %d개\n", coin, count)
		}
	}
}
