package main

import (
	"fmt"
)

func paintNeeded(width float32, length float32) (float32, error) {
	if width < 0 {
		return 0, fmt.Errorf("a width of %.2f is invalid", width)
	}
	if length < 0 {
		return 0, fmt.Errorf("a length of %.2f is invalid", length)
	}
	area := width * length
	//fmt.Printf("%.2f liter needed\n", area/10.0)
	return area / 10.0, nil
}

func printLiter(count float32, err error) {
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%.2f liter needed\n", count)
	}
}

func squares() func() int {
	var x int
	return func() int {
		x++
		return x * x
	}
}

func main() {
	//var num float64
	var num = 3.1415926
	var total, count float32

	//左对齐
	fmt.Printf("%-15s|%-10s\n", "Title", "Number")
	fmt.Printf("%-15s|%-5.5f\n", "No.1", num)

	count, err := paintNeeded(3.5, 4.8)
	printLiter(count, err)
	total += count

	count, err = paintNeeded(3.7, 8.8)
	printLiter(count, err)
	total += count

	count, err = paintNeeded(9.5, -2.8)
	printLiter(count, err)
	total += count

	fmt.Printf("%.2f total liter needed\n", total)

	f1 := squares()
	f2 := squares()
	fmt.Println("f1 1", f1())
	fmt.Println("f1 2", f1())
	fmt.Println("f2 1", f2())
	fmt.Println("f2 2", f2())
}
