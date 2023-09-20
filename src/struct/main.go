package main

import "fmt"

type subscriber struct {
	name   string
	rate   float64
	active bool
}

func printSubInfo(s subscriber) {
	fmt.Println(s.name, s.rate, s.active)
}

func defaultSubInfo(name string) subscriber {
	var s subscriber

	s.name = name
	s.active = true
	s.rate = 5.9

	return s
}

func main() {
	var sub1, sub2 subscriber
	sub1 = defaultSubInfo("allen")
	sub2 = defaultSubInfo("bell")
	sub2.rate = 4.3

	printSubInfo(sub1)
	// printSubInfo(sub2)
	printSubInfo(sub2)
}
