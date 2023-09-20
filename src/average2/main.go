// Package average2 is start from argvs
package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
)

func maximum(numbers ...float64) float64 {
	nums := numbers[0:]
	var tmpMax float64 = math.Inf(-1)

	for _, num := range nums {
		if num > tmpMax {
			tmpMax = num
		}
	}

	return tmpMax
}

func main() {
	//fmt.Println(len(os.Args), os.Args[1:])
	numbers := os.Args[1:]
	var maxNums []float64
	var sum float64

	for index, number := range numbers {
		fmt.Println("index:", index, "number:", number)
		number, err := strconv.ParseFloat(number, 64)
		if err != nil {
			log.Fatal(err)
		}
		maxNums = append(maxNums, number)
		sum += number
	}
	average := sum / float64(len(numbers))
	fmt.Printf("average: %.2f\n", average)

	fmt.Println(maximum(maxNums...))

	//fmt.Println(maximum(1.1, 2.2, 3.3, 4, 7, -3.2))
}
