package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Page struct {
	URL  string
	Size int
}

func reportNap(name string, delay int) {
	for i := 0; i < delay; i++ {
		fmt.Println(name, "sleeping")
		// 休眠1秒
		time.Sleep(1 * time.Second)
	}
	fmt.Println(name, "wakes up")
}

// 接收channel作为入参,传递的值为string
func send(myChannel chan string) {
	// 先休眠2s
	reportNap("send", 2)

	fmt.Println("Sending first value")
	myChannel <- "I'm first value"

	fmt.Println("Sending second value")
	myChannel <- "I'm second value"
}

func responseSize(url string, myChannel chan Page) {
	fmt.Println("Getting", url)
	respone, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer respone.Body.Close()

	body, err := io.ReadAll(respone.Body)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("url", url, "len", len(body))
	// myChannel <- len(body)
	myChannel <- Page{URL: url, Size: len(body)}
}

func main() {
	// myChannelWeb := make(chan int)
	myChannelWeb := make(chan Page)
	urls := []string{"https://www.baidu.com/", "https://www.baidu.com/", "https://www.baidu.com/", "https://www.baidu.com/",
		"https://www.163.com/", "https://www.163.com/", "https://www.163.com/", "https://www.163.com/",
		"https://www.qq.com/", "https://www.qq.com/", "https://www.qq.com/", "https://www.qq.com/"}
	// go responseSize("https://www.baidu.com/", myChannelWeb)
	// go responseSize("https://www.baidu.com/", myChannelWeb)
	// go responseSize("https://www.baidu.com/", myChannelWeb)
	// go responseSize("https://www.baidu.com/", myChannelWeb)
	// go responseSize("https://www.baidu.com/", myChannelWeb)
	// go responseSize("https://www.163.com/", myChannelWeb)
	// go responseSize("https://www.163.com/", myChannelWeb)
	// go responseSize("https://www.163.com/", myChannelWeb)
	// go responseSize("https://www.163.com/", myChannelWeb)
	// go responseSize("https://www.163.com/", myChannelWeb)
	// go responseSize("https://www.163.com/", myChannelWeb)
	// go responseSize("https://www.qq.com/", myChannelWeb)
	// go responseSize("https://www.qq.com/", myChannelWeb)
	// go responseSize("https://www.qq.com/", myChannelWeb)
	// go responseSize("https://www.qq.com/", myChannelWeb)
	// go responseSize("https://www.qq.com/", myChannelWeb)
	// go responseSize("https://www.qq.com/", myChannelWeb)
	// go responseSize("https://www.qq.com/", myChannelWeb)
	// go responseSize("https://www.qq.com/", myChannelWeb)
	for _, url := range urls {
		go responseSize(url, myChannelWeb)
	}

	for i := 0; i < len(urls); i++ {
		page := <-myChannelWeb
		fmt.Println(page.URL, page.Size)
	}

	// for _, url := range urls {
	// 	fmt.Println(url, <-myChannelWeb)
	// }

	// 接收与发送要配对，否则会出现deadlock
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)
	// fmt.Println(<-myChannelWeb)

	// 定义一个channel
	myChannel := make(chan string)

	// goroutine方式执行send
	go send(myChannel)

	// 休眠5s，保证上面的函数执行
	// time.Sleep(5 * time.Second)
	reportNap("main", 5)

	// fmt.Println("End")
	// 接收channel中第一个值，接收完第一个值，send中的第二个值才会被发送
	fmt.Println("receiving", <-myChannel)
	// 接收channel中第二个值
	fmt.Println("receiving", <-myChannel)

	// 打印的结果
	// 	main sleeping
	// send sleeping
	// send sleeping
	// main sleeping
	// main sleeping
	// send wakes up
	// Sending first value
	// main sleeping
	// main sleeping
	// main wakes up
	// receiving I'm first value
	// Sending second value
	// receiving I'm second value
}
