// 这是一个随机数的游戏
package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	//获取时间，将时间转换为整数
	seconds := time.Now()
	secondInt := time.Now().Unix()

	//rand.Seed(10)
	//不需要加Seed，新版go可以自动生成随机数
	target := rand.Intn(100) + 1

	fmt.Println("seconds:", seconds, "secondInt", secondInt)
	fmt.Println("Guess a number between 1-100")
	var success bool
	for i := 10; i > 0; i-- {
		fmt.Println("You can guess", i, "times")
		number := bufio.NewReader(os.Stdin)
		input, err := number.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Your enter is: %s", input)
		//去掉空格、换行等字符
		input = strings.TrimSpace(input)
		guess, err := strconv.Atoi(input)
		if err != nil {
			log.Fatal(err)
		}
		if guess > target {
			fmt.Println("Your guess is bigger than target")
		} else if guess < target {
			fmt.Println("Your guess is litter than target")
		} else {
			fmt.Println("Now, good, target is:", target, guess)
			success = true
			break
		}
	}
	if !success {
		fmt.Println("sorry, the target is:", target)
	}
}
