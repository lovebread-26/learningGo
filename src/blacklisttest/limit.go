package blacklist1

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis"
)

// 黑名单配置文件内容，如果扩展配置文件内容，记得修改这里
type ReqLimit struct {
	ReqLimit []ConfigReqLimit `json:"reqlimit"`
}

type ConfigReqLimit struct {
	CustId      string `json:"custId"`      // 客户ID
	MaxRequests int    `json:"maxRequests"` // 最大请求数
	Window      int    `json:"window"`      // 时间窗口
}

// 客户黑名单本地缓存数据结构
type cachelimitlist struct {
	data map[string]interface{}
	lock sync.RWMutex
}

type cacheItem struct {
	custId string
	token  string
	maxReq int
	window int
}

// 客户信息文件配置路径
var ReqLimitFilePath = "../../config/reqlimit.json"

// var config ConfigBlackList
var gConfig ReqLimit
var gRedis *redis.Client

// 客户黑名单本地缓存
var CacheReqLimit cachelimitlist

// 创建本地缓存
// 入参：无
// 返回值：cacheblacklist指针
func newCache() *cachelimitlist {
	// fmt.Println("Create a new cache")
	return &cachelimitlist{data: make(map[string]interface{})}
}

// 向本地缓存添加值
// 入参：键值key，值value
// 返回值：添加失败的原因
func (c *cachelimitlist) add(key string, value interface{}) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	// fmt.Println("Add key", key, "value", value)

	c.data[key] = value

	return nil
}

// 删除本地缓存中的值
// 入参：键值key
// 返回值：无
func (c *cachelimitlist) del(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	// fmt.Println("Del key", key)
	delete(c.data, key)
}

// 清空本地缓存
// 入参：无
// 返回值：无
func (c *cachelimitlist) clear() {
	c.lock.Lock()
	defer c.lock.Unlock()
	// fmt.Println("Clear the cache")
	c.data = make(map[string]interface{})
}

// 获取本地缓存值
// 入参：键值key
// 返回值：键值key对应的值value，键值是否存在ok
func (c *cachelimitlist) get(key string) (interface{}, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	value, ok := c.data[key]
	// fmt.Println("Get key", key, "value", value, "ok", ok)
	return value, ok
}

// 判断key是否在本地缓存中
// 入参：键值key
// 返回值：键值是否存在ok
func (c *cachelimitlist) contains(key string) bool {
	_, ok := c.get(key)
	// fmt.Println("Contains key", key, "ok", ok)
	return ok
}

// 更新客户ID黑名单本地缓存
// 入参：BlackList
// 返回值：无
func initCacheReqLimit(config ReqLimit) {
	for _, cfg := range config.ReqLimit {
		fmt.Println("init custId", cfg.CustId)
		item := cacheItem{custId: cfg.CustId, maxReq: cfg.MaxRequests, window: cfg.Window}
		CacheReqLimit.add(cfg.CustId, item)
	}
}

// 根据客户token查询本地缓存
// 入参：token
// 返回值，token是否存在
func SearchBlacklistByToken(token string) bool {
	return cacheToken.contains(token)
}

// 根据客户ID查询本地缓存
// 入参：客戶ID，客户token
// 返回值，客戶ID是否存在
func SearchBlacklistByCustId(custId string, token string) bool {
	ok := cacheCustId.contains(custId)
	if ok {
		updateCacheToken(token, custId)
	}
	return ok
}

const fixedWindowLimiterTryAcquireRedisScript = `
-- ARGV[1]: 窗口时间大小
-- ARGV[2]: 窗口请求上限

local window = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])

-- 获取原始值
local counter = tonumber(redis.call("get", KEYS[1]))
if counter == nil then 
   counter = 0
end
-- 若到达窗口请求上限，请求失败
if counter >= limit then
   return 0
end
-- 窗口值+1
redis.call("incr", KEYS[1])
if counter == 0 then
    redis.call("pexpire", KEYS[1], window)
end
return 1
`

// FixedWindowLimiter 固定窗口限流器
type FixedWindowLimiter struct {
	limit  int           // 窗口请求上限
	window int           // 窗口时间大小
	client *redis.Client // Redis客户端
	script *redis.Script // TryAcquire脚本
}

func NewFixedWindowLimiter(client *redis.Client, limit int, window time.Duration) (*FixedWindowLimiter, error) {
	return &FixedWindowLimiter{
		limit:  limit,
		window: int(window / time.Millisecond),
		client: client,
		script: redis.NewScript(fixedWindowLimiterTryAcquireRedisScript),
	}, nil
}

func (l *FixedWindowLimiter) TryAcquire(ctx context.Context, resource string) error {
	success, err := l.script.Run(ctx, l.client, []string{resource}, l.window, l.limit).Bool()
	if err != nil {
		return err
	}
	// 若到达窗口请求上限，请求失败
	if !success {
		return ErrAcquireFailed
	}
	return nil
}

// 加载配置文件，程序启动时加载。配置文件被修改后，手动通过调用接口加载
func LoadConfigReqLimitFile(filePath string) {
	// // 清除本地缓存，无论配置文件是否加载成功
	CacheReqLimit.clear()
	// CacheToken.Clear()

	// 读取json文件
	jsonData, err := os.ReadFile(filePath)
	CheckErr(err)

	// 解析json内容
	err = json.Unmarshal(jsonData, &gConfig)
	CheckErr(err)

	// fmt.Printf("%#v\n", config)
	for _, cfg := range gConfig.ReqLimit {
		fmt.Printf("%#v\n", cfg)
	}

	// 初始化客戶ID本地緩存
	initCacheReqLimit(gConfig)
}

// 打印配置文件内容
// 入參：无
// 返回值：无
func PrintReqLimit() {
	fmt.Printf("================================ 客户ID限速配置 =================================\n")
	fmt.Printf("|%-38s|%-14s|%-19s|\n", "客户ID", "最大请求", "Window")
	for _, value := range gConfig.ReqLimit {

		fmt.Printf("---------------------------------------------------------------------------------\n")
		fmt.Printf("|%-40s|%-18d|%-19d|\n", value.CustId, value.MaxRequests, value.Window)

	}
	fmt.Printf("=================================================================================\n")
}

func InitReqLimit(redis *redis.Client) {
	// 创建本地缓存
	CacheReqLimit = *newCache()
	// CacheToken = *NewCache()

	gRedis = redis

	// 加载配置文件,test
	LoadConfigReqLimitFile(ReqLimitFilePath)

	// http.HandleFunc("/reloadBlacklist", handleReloadConfig)
	// http.HandleFunc("/printCacheBlacklist", handlePrintCacheBlacklist)

	// 监听8080端口
	// err := http.ListenAndServe(":8080", nil)

	// 下面两行如果不是出错，不会执行
	// log.Fatal(err)
	PrintReqLimit()
	// time.Sleep(5 * time.Second)
	// CacheCustId.Clear()
	// CacheToken.Clear()
	// PrintCacheBlacklist()
}
