// 客户限速管理软件包
// 根据客户请求的token和限速配置文件中的客户ID，判断一个客户请求是否超过每秒允许的最大值
// 如果客户请求超过每秒允许最大值，拒绝请求的后续处理
// 下一秒恢复请求最大值
package blacklist1

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"golang.org/x/time/rate"
)

// 客户限速配置文件内容，如果扩展配置文件内容，记得修改这里
type reqLimit struct {
	ReqLimit []configReqLimit `json:"reqlimit"`
}

type configReqLimit struct {
	CustId      string `json:"custId"`      // 客户ID
	MaxRequests int    `json:"maxRequests"` // 每秒钟最大请求数
	// Window      int    `json:"window"`      // 时间窗口，秒
}

// 客户限速本地缓存数据结构
type cacheReqLimit struct {
	data map[string]interface{}
	lock sync.RWMutex
}

type cacheItem struct {
	custId string
	token  string
	maxReq int
	// window  int
	limiter *rate.Limiter
}

type InitConfigReqLimit struct {
	ReqLimitFilePath string // 配置文件名
}

// 客户限速本地缓存
var CacheCustId cacheReqLimit
var CacheToken cacheReqLimit
var gConfig InitConfigReqLimit
var ReqLimitFilePath = "../../config/reqlimit.json"
var gRedis *redis.Client
var reqCount int

// 创建本地缓存
// 入参：无
// 返回值：cacheReqLimit指针
func newCache() *cacheReqLimit {
	// fmt.Println("Create a new cache")
	return &cacheReqLimit{data: make(map[string]interface{})}
}

// 向本地缓存添加值
// 入参：键值key，值value
// 返回值：添加失败的原因
func (c *cacheReqLimit) add(key string, value interface{}) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	// fmt.Println("Add key", key, "value", value)

	c.data[key] = value

	return nil
}

// 删除本地缓存中的值
// 入参：键值key
// 返回值：无
func (c *cacheReqLimit) del(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	// fmt.Println("Del key", key)
	delete(c.data, key)
}

// 清空本地缓存
// 入参：无
// 返回值：无
func (c *cacheReqLimit) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()
	// fmt.Println("Clear the cache")
	c.data = make(map[string]interface{})
}

// 获取本地缓存值
// 入参：键值key
// 返回值：键值key对应的值value，键值是否存在ok
func (c *cacheReqLimit) get(key string) (interface{}, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	value, ok := c.data[key]
	// fmt.Println("Get key", key, "value", value, "ok", ok)
	return value, ok
}

// 判断key是否在本地缓存中
// 入参：键值key
// 返回值：键值是否存在ok，键值对应的值interface{}
func (c *cacheReqLimit) contains(key string) (bool, interface{}) {
	item, ok := c.get(key)
	fmt.Println("Contains key", key, "ok", ok)
	return ok, item
}

// 更新客户ID限速本地缓存
// 入参：reqLimit
// 返回值：无
func initCacheCustId(config reqLimit) {
	for _, cfg := range config.ReqLimit {
		fmt.Println("init custId", cfg.CustId)
		// 每秒放置MaxRequests个令牌，最多存储MaxRequests个令牌
		limiter := rate.NewLimiter(rate.Limit(float64(cfg.MaxRequests)), cfg.MaxRequests)
		item := cacheItem{custId: cfg.CustId, maxReq: cfg.MaxRequests, limiter: limiter}
		CacheCustId.add(cfg.CustId, item)
	}
}

// 删除过期token
// 入参：客户ID
// 返回值：无
func removeCacheTokenByCustId(custId string) {
	value, ok := CacheCustId.get(custId)
	if ok {
		item, ok := value.(cacheItem)
		if ok {
			// fmt.Println("remove custId", custId, "token", item.token)
			CacheToken.del(item.token)
		}
	}
}

// 更新客户ID限速本地缓存
// 入参：客户token, 客户ID custId
// 返回值：无
func updateCacheCustId(token string, custId string) cacheItem {
	// fmt.Println("update token", token, "custId", custId)
	items := cacheItem{custId: custId, token: token}
	value, ok := CacheCustId.get(custId)
	if ok {
		item, ok := value.(cacheItem)
		if ok {
			items = cacheItem{custId: custId, token: token, maxReq: item.maxReq, limiter: item.limiter}
			CacheCustId.add(custId, items)
		}
	}

	return items
}

// 更新客户token限速本地缓存
// 入参：客户token, 客户ID custId
// 返回值：无
func updateCacheToken(token string, custId string) {
	// fmt.Println("update token", token, "custId", custId)

	// item := cacheItem{custId: custId, token: token}

	removeCacheTokenByCustId(custId)

	item := updateCacheCustId(token, custId)

	CacheToken.add(token, item)
}

func handelLimitByToken(item cacheItem) {
	limiter := item.limiter

	if limiter != nil {
		reqCount++
		now := time.Now()
		formattedTime := now.Format("2006-01-02 15:04:05")
		if limiter.Allow() {
			fmt.Println("现在是：", formattedTime, "第", reqCount, "个请求，不限速", "令牌数是", limiter.Tokens())
		} else {
			fmt.Println("现在是：", formattedTime, "第", reqCount, "个请求，限速", "令牌数是", limiter.Tokens())
		}
	}
}

// 根据客户token查询本地缓存
// 入参：token
// 返回值，token是否存在
func SearchReqLimitByToken(token string) bool {
	ok, value := CacheToken.contains(token)
	if ok {
		// 如果找到了，调用限速处理
		item, ok := value.(cacheItem)
		if ok {
			handelLimitByToken(item)
		}
	}
	return ok
}

// 根据客户ID查询本地缓存，如果找到客户ID，去更新客户ID对应的token本地缓存
// 入参：客戶ID，客户token
// 返回值，客戶ID是否存在
func SearchReqLimitByCustId(custId string, token string) bool {
	ok, _ := CacheCustId.contains(custId)
	if ok {
		updateCacheToken(token, custId)
	}
	return ok
}

// 打印本地緩存內容
// 入參：无
// 返回值：无
// curl "http://127.0.0.1:8080/reqlimitlist?print=1"
func PrintCacheReqLimit() {
	CacheCustId.lock.RLock()
	defer CacheCustId.lock.RUnlock()
	fmt.Printf("============================================ 客户ID本地缓存 ===========================================\n")
	fmt.Printf("|%-38s|%-38s|%-20s|\n", "客户ID", "Token", "每秒最大请求数")
	for key, value := range CacheCustId.data {
		item, ok := value.(cacheItem)
		if ok {
			fmt.Printf("-------------------------------------------------------------------------------------------------------\n")
			fmt.Printf("|%-40s|%-38s|%-20d|\n", key, item.token, item.maxReq)
		}
	}
	fmt.Printf("=======================================================================================================\n")

	CacheToken.lock.RLock()
	defer CacheToken.lock.RUnlock()
	fmt.Printf("========================================== 客户Token本地缓存 ==========================================\n")
	fmt.Printf("|%-38s|%-38s|%-20s|\n", "客户Token", "CustId", "每秒钟最大请求数")
	for key, value := range CacheToken.data {
		item, ok := value.(cacheItem)
		if ok {
			fmt.Printf("-----------------------------------------------------------------------------------------------------\n")
			fmt.Printf("|%-40s|%-38s|%-20d|\n", key, item.custId, item.maxReq)
		}
	}
	fmt.Printf("=======================================================================================================\n")
}

// 统一的错误处理
func CheckErr(err error) {
	if err != nil {
		// fmt.Println(err)
		log.Fatal(err)
	}
}

// 加载配置文件，程序启动时加载
// 配置文件被修改后，手动通过调用接口加载
func LoadConfigReqLimitFile(filePath string) {
	// 清除本地缓存，无论配置文件是否加载成功
	CacheCustId.Clear()
	CacheToken.Clear()

	// 读取json文件
	jsondata, err := os.ReadFile(filePath)
	CheckErr(err)

	// 解析json内容
	// var config ConfigReqReqLimit
	var config reqLimit
	err = json.Unmarshal(jsondata, &config)
	CheckErr(err)

	// fmt.Printf("%#v\n", config)
	for _, cfg := range config.ReqLimit {
		fmt.Printf("%#v\n", cfg)
	}

	// 初始化客戶ID本地緩存
	initCacheCustId(config)
}

// 客户限速接口处理函数
// 入参：请求内容
// 返回值：无
func ReqLimitHandleRequest(c *gin.Context) {
	// 获取查询参数（如果存在）
	queryParams := c.Request.URL.Query()
	// fmt.Println("Query parameters:", queryParams)

	// 获取参数值
	print := queryParams.Get("reqlimitlistprint")
	reload := queryParams.Get("reqlimitlistreload")
	Clear := queryParams.Get("reqlimitlistClear")

	// 打印本地缓存内容
	if print == "1" {
		PrintCacheReqLimit()
	}

	// 重新加载配置文件
	if reload == "1" {
		LoadConfigReqLimitFile(gConfig.ReqLimitFilePath)
	}

	// 清除本地缓存
	// curl "http://127.0.0.1:8080/reqlimitlist?Clear=1"
	if Clear == "1" {
		CacheCustId.Clear()
		CacheToken.Clear()
	}
}

// 初始化配置文件
// 入参：config，本地缓存的配置
// 返回值：无
func initConfig() {
	// gConfig.ReqLimitFilePath = config.ReqLimitFilePath
	// if len(gConfig.ReqLimitFilePath) == 0 {
	// 	gConfig.ReqLimitFilePath = "reqlimit.json"
	// }

	// 加载配置文件
	LoadConfigReqLimitFile(ReqLimitFilePath)
}

// 客户限速初始化
// 入参：redis连接
// 返回值：无
func InitReqLimit(redis *redis.Client) {
	// 创建本地缓存
	CacheCustId = *newCache()
	CacheToken = *newCache()

	gRedis = redis

	// 初始化全局配置
	initConfig()

	PrintCacheReqLimit()
}
