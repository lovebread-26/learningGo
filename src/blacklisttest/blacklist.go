// 客户黑名单管理软件包。
// 根据客户请求的token和黑名单配置文件中的客户ID，判断一个客户请求是否在黑名单中。
// 如果请求在黑名单中，拒绝请求的后续处理；如果请求不在黑名单中，允许请求的后续处理。
package blacklist1

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
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

type CacheItem struct {
	CustId     string
	Token      string
	ExpireTime time.Time
}

// 客户黑名单本地缓存
var CacheCustId CacheBlacklist
var CacheToken CacheBlacklist

// 黑名单文件配置路径
var BlacklistFilePath = "../../config/客户黑名单配置文件1.json"

// 客户token本地缓存过期时间，2小时
// var tokenExpireTime = 2 * time.Hour
var tokenExpireTime = 1 * time.Minute

// token本地缓存上限为10000
// var tokenMaxSize = 10000
var tokenMaxSize = 2

// 创建本地缓存
// 入参：无
// 返回值：CacheBlacklist指针
func NewCache() *CacheBlacklist {
	fmt.Println("Create a new cache")
	return &CacheBlacklist{Data: make(map[string]interface{})}
}

// 向本地缓存添加值
// 入参：键值key，值value，过期时间：如果过期时间为0，则不设置过期时间
// 返回值：添加失败的原因
func (c *CacheBlacklist) Add(key string, value interface{}) error {
	if len(c.Data) >= tokenMaxSize {
		fmt.Println("map is full,max:", tokenMaxSize, "now:", len(c.Data))
		return fmt.Errorf("map is full")
	}

	c.Lock.Lock()
	defer c.Lock.Unlock()
	fmt.Println("Add key", key, "value", value)

	c.Data[key] = value

	return nil
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
		item := CacheItem{CustId: cfg.CustId}
		CacheCustId.Add(cfg.CustId, item)
	}
}

// 并判断是否过期,如果过期则删除
func removeExpiredToken() {
	for key, value := range CacheToken.Data {
		item, ok := value.(CacheItem)
		if ok {
			// fmt.Println("now:", time.Now().Format("2006-01-02 15:04:05"), "expiretime:", item.ExpireTime.Format("2006-01-02 15:04:05"))
			if time.Now().After(item.ExpireTime) {
				CacheToken.Del(key)
			}
		}
	}
}

// 更新客户token黑名单本地缓存
// 入参：token, custId
// 返回值：无
func updateCacheToken(token string, custId string) {
	fmt.Println("update token", token, "custId", custId)

	// 更新前先清理过期的token
	removeExpiredToken()

	// 获取当前时间
	currentTime := time.Now()
	// 计算过期时间
	var config int64
	var expire time.Duration
	config = 60
	expire = time.Duration(config)
	futureTime := currentTime.Add(expire)

	item := CacheItem{CustId: custId, Token: token, ExpireTime: futureTime}

	CacheToken.Add(token, item)
}

// 根据客户token查询本地缓存
// 入参：token, redis
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
		updateCacheToken(token, custId)
	}
	return ok
}

// 打印本地緩存內容
// 入參：无
// 返回值：无
func PrintCacheBlacklist() {
	fmt.Printf("================================ 客户ID本地缓存 =================================\n")
	fmt.Printf("|%-38s|%-38s|\n", "客户ID", "Token")
	for key, value := range CacheCustId.Data {
		item, ok := value.(CacheItem)
		if ok {
			fmt.Printf("---------------------------------------------------------------------------------\n")
			fmt.Printf("|%-40s|%-38s|\n", key, item.Token)
		}
	}
	fmt.Printf("=================================================================================\n")

	fmt.Printf("========================================= 客户Token本地缓存 =========================================\n")
	fmt.Printf("|%-38s|%-38s|%-15s|\n", "客户Token", "CustId", "过期时间")
	for key, value := range CacheToken.Data {
		item, ok := value.(CacheItem)
		if ok {
			fmt.Printf("-----------------------------------------------------------------------------------------------------\n")
			fmt.Printf("|%-40s|%-38s|%-15s|\n", key, item.CustId, item.ExpireTime.Format("2006-01-02 15:04:05"))
		}
	}
	fmt.Printf("=====================================================================================================\n")
}

// 统一的错误处理函数
func CheckErr(err error) {
	if err != nil {
		// log.Fatal(err)
		fmt.Println(err)
	}
}

// 加载配置文件，程序启动时加载。配置文件被修改后，手动通过调用接口加载
func LoadConfigBlacklistFile(filePath string) {
	// 清除本地缓存，无论配置文件是否加载成功
	CacheCustId.Clear()
	CacheToken.Clear()

	// 读取json文件
	jsonData, err := os.ReadFile(filePath)
	CheckErr(err)

	// 解析json内容
	// var config ConfigBlackList
	var config BlackList
	err = json.Unmarshal(jsonData, &config)
	CheckErr(err)

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
