// 数组处理平均值
package main

import (
	"fmt"
	"log"
	"readfloat"
)

func main() {
	//numbers := [3]float64{71.8, 56.2, 89.5}
	numbers, err := readfloat.ReadFloatFromFile("../../config/data1.txt")
	if err != nil {
		log.Fatal(err)
	}
	var sum float64 = 0
	fmt.Printf("numbers:%#v\n", numbers)
	for index, num := range numbers {
		fmt.Println("index:", index, "number:", num)
		sum += num
	}
	fmt.Println("total:", sum)

	average := sum / float64(len(numbers))
	fmt.Printf("average is %.2f\n", average)
}
