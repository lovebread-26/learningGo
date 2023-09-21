package main

import "fmt"

type Date struct {
	Year  int
	Month int
	Day   int
}

func main() {
	day := Date{Year: 2023, Month: 9, Day: 21}

	fmt.Println(day)
}
