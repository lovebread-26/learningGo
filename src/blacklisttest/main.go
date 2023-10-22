// 黑名单测试程序
package main

import (
	"blacklist"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func handleCustId(w http.ResponseWriter, r *http.Request) {
	// 解析URL中的查询参数
	params, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		log.Fatal(err)
	}

	// 获取参数值
	custid := params.Get("custid")
	token := params.Get("token")
	print := params.Get("print")
	reload := params.Get("reload")
	clear := params.Get("clear")

	// 打印参数值
	fmt.Println("request custid:", custid)
	fmt.Println("request token:", token)

	if ok := blacklist.SearchBlacklistByToken(token); ok {
		fmt.Println("token", token, "in blacklist")
	} else if ok := blacklist.SearchBlacklistByCustId(custid, token); ok {
		fmt.Println("custId", custid, "in blacklist")
		// blacklist.updateCacheToken(custid)
	} else {
		fmt.Println("custId", custid, "and token", token, "don't in blacklist")
	}

	if print == "1" {
		blacklist.PrintCacheBlacklist()
	}

	if reload == "1" {
		blacklist.LoadConfigBlacklistFile(blacklist.BlacklistFilePath)
	}

	if clear == "1" {
		blacklist.CacheCustId.Clear()
		blacklist.CacheToken.Clear()
	}
}

func handleRequest(c *gin.Context) {
	// 处理请求逻辑
	// 根据请求类型和路径执行相应的操作

	// 获取请求方法（GET、POST等）
	method := c.Request.Method
	fmt.Println("Request method:", method)

	// 获取请求路径
	path := c.Request.URL.Path
	fmt.Println("Request path:", path)

	// 获取查询参数（如果存在）
	queryParams := c.Request.URL.Query()
	fmt.Println("Query parameters:", queryParams)

	// 获取请求体（如果请求方法为POST或PUT）
	// requestBody := c.Request.Body
	// fmt.Println("Request body:", requestBody)

	// 设置响应状态码和响应体
	// c.Status(http.StatusOK)
	// fmt.Println(c, "Hello, World!")

	// 获取参数值
	custid := queryParams.Get("custid")
	token := queryParams.Get("token")
	print := queryParams.Get("print")
	reload := queryParams.Get("reload")
	clear := queryParams.Get("clear")

	// 打印参数值
	fmt.Println("request custid:", custid)
	fmt.Println("request token:", token)

	if ok := blacklist.SearchBlacklistByToken(token); ok {
		fmt.Println("token", token, "in blacklist")
	} else if ok := blacklist.SearchBlacklistByCustId(custid, token); ok {
		fmt.Println("custId", custid, "in blacklist")
		// blacklist.updateCacheToken(custid)
	} else {
		fmt.Println("custId", custid, "and token", token, "don't in blacklist")
	}

	if print == "1" {
		blacklist.PrintCacheBlacklist()
	}

	if reload == "1" {
		blacklist.LoadConfigBlacklistFile(blacklist.BlacklistFilePath)
	}

	if clear == "1" {
		blacklist.CacheCustId.Clear()
		blacklist.CacheToken.Clear()
	}
}

func main() {
	// 黑名单初始化
	blacklist.Initblacklist()

	// 创建一个gin的路由器实例
	router := gin.Default()

	// 使用Any方法定义一个处理函数，处理所有类型的HTTP请求
	router.Any("/blacklist", handleRequest)

	// 启动服务器并监听端口
	router.Run(":8080")

	// http.HandleFunc("/blacklist", handleCustId)

	// 启动HTTP服务器
	// err := http.ListenAndServe(":8080", nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
