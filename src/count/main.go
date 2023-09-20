package main

import (
	"fmt"
	"log"
	"readfloat"
	"sort"
)

func main() {
	var count []int
	var names []string
	counts := make(map[string]int)

	strings, err := readfloat.ReadStringFromFile("../../config/data2.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings, len(strings))

	for _, name := range strings {
		matched := false
		for i, line := range names {
			if line == name {
				count[i]++
				matched = true
			}
		}

		if !matched {
			names = append(names, name)
			count = append(count, 1)
		}

		counts[name]++
	}

	//名字按字母表排序
	sort.Strings(names)

	fmt.Printf("%-10s | %-10s\n", "name", "count")
	for i, name := range names {
		fmt.Printf("%-10s | %-10d\n", name, count[i])
	}

	fmt.Println("----------------")
	for i, name := range counts {
		// fmt.Println(i, name)
		fmt.Printf("%-10s | %-10d\n", i, name)
	}

	fmt.Println("----------------")
	for _, name := range names {
		fmt.Printf("%-10s | %-10d\n", name, counts[name])
	}
}
