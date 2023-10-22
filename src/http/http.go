package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

type Guestbook struct {
	SignatureCount int
	Signatures     []string
}

// 获取文本内容
func getStrings(fileName string) []string {
	var lines []string
	file, err := os.Open(fileName)
	if os.IsNotExist(err) {
		return nil
	}
	check(err)
	// 保证文件被关闭
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	check(scanner.Err())
	return lines
}

// 错误检查
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func write(w http.ResponseWriter, msg string) {
	// message := []byte(msg)
	// _, err := w.Write(message)
	// check(err)
	signature := getStrings("signature.txt")
	// fmt.Printf("%#v\n", signature)
	signatures := Guestbook{SignatureCount: len(signature), Signatures: signature}

	html, err := template.ParseFiles(msg)
	check(err)
	// err = html.Execute(w, nil)
	err = html.Execute(w, signatures)
	check(err)
}

func aHandler(response http.ResponseWriter, request *http.Request) {
	write(response, "view.html")
}

func bHandler(response http.ResponseWriter, request *http.Request) {
	write(response, "new.html")
}

func hellobCreateHandler(response http.ResponseWriter, request *http.Request) {
	// write(response, "hello web c!")
	signature := request.FormValue("signature")
	// _, err := response.Write([]byte(signature))
	// fmt.Println("signature:", signature)
	// check(err)
	options := os.O_WRONLY | os.O_APPEND | os.O_CREATE
	file, err := os.OpenFile("signature.txt", options, os.FileMode(0600))
	check(err)
	_, err = fmt.Fprintln(file, signature)
	check(err)
	check(file.Close())

	// 重定向
	http.Redirect(response, request, "/helloa", http.StatusFound)
}

func main() {
	// text := "This is a text template!\n"
	// action
	text := "This is a text template!\nAction:{{.}}\n"
	tmpl, err := template.New("test").Parse(text)
	check(err)
	// 输出到终端
	err = tmpl.Execute(os.Stdout, "templates action")
	check(err)

	// 添加监控的请求及响应的函数，收到/hello request，调用viewHandler函数来响应
	http.HandleFunc("/helloa", aHandler)
	http.HandleFunc("/hellob", bHandler)
	http.HandleFunc("/hellob/create", hellobCreateHandler)

	// 监听8080端口
	err = http.ListenAndServe(":8080", nil)

	// 下面两行如果不是出错，不会执行
	fmt.Println("Can you see me?")
	log.Fatal(err)
}
