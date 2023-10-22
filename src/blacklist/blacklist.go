// 客户黑名单管理软件包。
// 根据客户请求的token和黑名单配置文件中的客户ID，判断一个客户请求是否在黑名单中。
// 如果请求在黑名单中，拒绝请求的后续处理；如果请求不在黑名单中，允许请求的后续处理。
package blacklist

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

// 黑名单配置文件内容，如果扩展配置文件内容，记得修改这里
type BlackList struct {
	BlackList []ConfigBlackList `json:"blacklist"`
}

type ConfigBlackList struct {
	CustId       string `json:"custId"`       // 客户ID
	OrderId      string `json:"orderId"`      // 订单ID
	CustName     string `json:"custName"`     // 客户名称
	Province     string `json:"province"`     // 业务发展省
	ManagerName  string `json:"managerName"`  // 客户经理名称
	ManagerPhone string `json:"managerPhone"` // 客户经理电话
}

// 客户黑名单本地缓存数据结构
type CacheBlacklist struct {
	Data map[string]interface{}
	Lock sync.RWMutex
}

// 客户黑名单本地缓存
var CacheCustId CacheBlacklist
var CacheToken CacheBlacklist

// 黑名单文件配置路径
var BlacklistFilePath = "../../config/客户黑名单配置文件1.json"

// 创建本地缓存
// 入参：无
// 返回值：CacheBlacklist指针
func NewCache() *CacheBlacklist {
	fmt.Println("Create a new cache")
	return &CacheBlacklist{Data: make(map[string]interface{})}
}

// 向本地缓存添加值
// 入参：键值key，值value
// 返回值：无
func (c *CacheBlacklist) Add(key string, value interface{}) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	fmt.Println("Add key", key, "value", value)
	c.Data[key] = value
}

// 删除本地缓存中的值
// 入参：键值key
// 返回值：无
func (c *CacheBlacklist) Del(key string) {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	fmt.Println("Del key", key)
	delete(c.Data, key)
}

// 清空本地缓存
// 入参：无
// 返回值：无
func (c *CacheBlacklist) Clear() {
	c.Lock.Lock()
	defer c.Lock.Unlock()
	fmt.Println("Clear the cache")
	c.Data = make(map[string]interface{})
}

// 获取本地缓存值
// 入参：键值key
// 返回值：键值key对应的值value，键值是否存在
func (c *CacheBlacklist) Get(key string) (interface{}, bool) {
	c.Lock.RLock()
	defer c.Lock.RUnlock()
	value, ok := c.Data[key]
	fmt.Println("Get key", key, "value", value, "ok", ok)
	return value, ok
}

// 判断key是否在本地缓存中
// 入参：键值key
// 返回值：键值是否存在
func (c *CacheBlacklist) Contains(key string) bool {
	_, ok := c.Get(key)
	fmt.Println("Contains key", key, "ok", ok)
	return ok
}

// 更新客户ID黑名单本地缓存
// 入参：BlackList
// 返回值：无
func initCacheCustId(config BlackList) {
	for _, cfg := range config.BlackList {
		fmt.Println("init custId", cfg.CustId)
		CacheCustId.Add(cfg.CustId, true)
	}
}

// 更新客户token黑名单本地缓存
// 入参：token
// 返回值：无
func updateCacheToken(token string) {
	fmt.Println("update token", token)
	CacheToken.Add(token, true)
}

// 根据客户token查询本地缓存
// 入参：token
// 返回值，token是否存在
func SearchBlacklistByToken(token string) bool {
	return CacheToken.Contains(token)
}

// 根据客户ID查询本地缓存
// 入参：客戶ID，客户token
// 返回值，客戶ID是否存在
func SearchBlacklistByCustId(custId string, token string) bool {
	ok := CacheCustId.Contains(custId)
	if ok {
		updateCacheToken(token)
	}
	return ok
}

// 打印本地緩存內容
// 入參：无
// 返回值：无
func PrintCacheBlacklist() {
	fmt.Printf("================== 客户ID本地缓存 ===================\n")
	fmt.Printf("|%-38s|%-9s|\n", "客户ID", "值")
	for key, value := range CacheCustId.Data {
		fmt.Printf("-----------------------------------------------------\n")
		fmt.Printf("|%-40s|%-10t|\n", key, value)
	}
	fmt.Printf("=====================================================\n")

	fmt.Printf("================== 客户Token本地缓存 ================\n")
	fmt.Printf("|%-38s|%-9s|\n", "客户Token", "值")
	for key, value := range CacheToken.Data {
		fmt.Printf("-----------------------------------------------------\n")
		fmt.Printf("|%-40s|%-10t|\n", key, value)
	}
	fmt.Printf("=====================================================\n")
}

// 统一的错误处理函数
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// 加载配置文件，程序启动时加载。配置文件被修改后，手动通过调用接口加载
func LoadConfigBlacklistFile(filePath string) {
	// 清除本地缓存，无论配置文件是否加载成功
	CacheCustId.Clear()
	CacheToken.Clear()

	// 读取json文件
	jsonData, err := os.ReadFile(filePath)
	checkErr(err)

	// 解析json内容
	// var config ConfigBlackList
	var config BlackList
	err = json.Unmarshal(jsonData, &config)
	checkErr(err)

	// fmt.Printf("%#v\n", config)
	for _, cfg := range config.BlackList {
		fmt.Printf("%#v\n", cfg)
	}

	// 初始化客戶ID本地緩存
	initCacheCustId(config)
}

// 接口处理函数
func handleReloadConfig(w http.ResponseWriter, r *http.Request) {
	LoadConfigBlacklistFile(BlacklistFilePath)
}
func handlePrintCacheBlacklist(w http.ResponseWriter, r *http.Request) {
	PrintCacheBlacklist()
}

func Initblacklist() {
	// 创建本地缓存
	CacheCustId = *NewCache()
	CacheToken = *NewCache()

	// 加载配置文件,test
	LoadConfigBlacklistFile(BlacklistFilePath)

	http.HandleFunc("/reloadBlacklist", handleReloadConfig)
	http.HandleFunc("/printCacheBlacklist", handlePrintCacheBlacklist)

	// 监听8080端口
	// err := http.ListenAndServe(":8080", nil)

	// 下面两行如果不是出错，不会执行
	// log.Fatal(err)
	PrintCacheBlacklist()
	// time.Sleep(5 * time.Second)
	// CacheCustId.Clear()
	// CacheToken.Clear()
	// PrintCacheBlacklist()
}
