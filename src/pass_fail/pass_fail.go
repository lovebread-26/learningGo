// 这是一个学习成绩查询文件
package main

import (
	"fmt"
	"keyboard"
	"log"
)

func main() {
	fmt.Print("Enter a grade: ")
	/*
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		input = strings.TrimSpace(input)
		//fmt.Println(input)

		grade, err := strconv.ParseFloat(input, 64)
		if err != nil {
			log.Fatal(err)
		}
	*/
	var status string
	grade, err := keyboard.Getfloat()
	if err != nil {
		log.Fatal(err)
	}
	if grade >= 60 {
		status = "passing"
	} else {
		status = "failing"
	}

	fmt.Println("A grade of", grade, "is", status)
}
