// 黑名单测试程序
package main

import (
	"blacklist1"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// redis服务器
var redisAddr = "192.168.225.128:6379"
var gRedis *redis.Client
var Ctx context.Context
var gCollection *mongo.Collection
var UtilsMongoCol unsafe.Pointer

type logMsg struct {
	user string
	age  int
}

// 客户信息文件配置路径
var CustInfoFilePath = "../../config/客户信息配置文件.json"

// 客户信息数据格式
type CustInfo struct {
	Key      string `json:"key"`
	ExpireAt int    `json:"expireAt"`
	Data     Data   `json:"data"`
}

type Data struct {
	ID              string `json:"id"`
	Data            Data1  `json:"data"`
	IsAdmin         bool   `json:"isAdmin"`
	AllowMultiLogin bool   `json:"allowMultiLogin"`
}

type Data1 struct {
	ID            string `json:"_id"`
	Rid           int    `json:"rid"`
	Nickname      string `json:"nickname"`
	Source        string `json:"source"`
	CustID        string `json:"cust_id"`
	CustomerID    string `json:"customer_id"`
	Account       string `json:"account"`
	UID           string `json:"uid"`
	CreateTime    string `json:"create_time"`
	Status        string `json:"status"`
	AuthLockExp   int    `json:"authLockExp"`
	LastLoginIP   string `json:"lastLoginIp"`
	LastLoginTime int64  `json:"lastLoginTime"`
	PwdErrNum     int    `json:"pwdErrNum"`
	PwdChangeTime string `json:"pwd_change_time"`
	Apply         Apply  `json:"apply"`
}

type Apply struct {
	App       string `json:"app"`
	AppID     string `json:"app_id"`
	AppKey    string `json:"app_key"`
	AppSecret string `json:"app_secret"`
}

// 加载客户信息配置文件
// 加载配置文件，程序启动时加载。配置文件被修改后，手动通过调用接口加载
func LoadCustInfoFile(filePath string) {
	// 读取json文件
	jsonData, err := os.ReadFile(filePath)
	blacklist1.CheckErr(err)

	// 解析json内容
	var config CustInfo
	err = json.Unmarshal(jsonData, &config)
	blacklist1.CheckErr(err)

	// 将数据存储到redis
	fmt.Println("key", config.Key)
	gRedis.Set(config.Key, jsonData, 0)
}

// 获取客户信息
func getCustInfo(token string) CustInfo {
	// akey := "dev-op:AccessToken" + at
	key := strings.Join([]string{"dev-op:AccessToken", token}, ":")
	fmt.Println("redis key:", key)
	value, err := gRedis.Get(key).Result()
	blacklist1.CheckErr(err)
	// fmt.Printf("value:%#v\n", value)

	// 解析json
	var custInfo CustInfo
	err = json.Unmarshal([]byte(value), &custInfo)
	blacklist1.CheckErr(err)
	fmt.Println("custid", custInfo.Data.Data.CustID)
	fmt.Println("custInfo", custInfo)
	return custInfo
}

// 订阅redis消息
func handleRedisChannel() {
	// 订阅Redis频道
	channel := "gw_bind_back_go"
	pubsub := gRedis.PSubscribe(channel)
	// defer gRedis.Unsubscribe(channel)
	defer pubsub.Close()

	_, err := pubsub.Receive()
	if err != nil {
		log.Fatal(err)
	}

	// 等待消息
	ch := pubsub.Channel()
	for msg := range ch {
		fmt.Println(msg.Channel, msg.Pattern, msg.Payload)
	}
}

// 连接redis
// 入参：无。
// 返回值：无。
func HandleRedis() {
	options := redis.Options{
		Network: "tcp",
		Addr:    redisAddr,
		// Dialer:             (func() (net.Conn, error))(0xc210c0),
		// OnConnect:          (func(*redis.Conn) error)(0xc248a0),
		Password:           "",
		DB:                 0,
		MaxRetries:         0,
		MinRetryBackoff:    8000000,
		MaxRetryBackoff:    512000,
		DialTimeout:        5000000000,
		ReadTimeout:        500000000,
		WriteTimeout:       3000000000,
		PoolSize:           15,
		MinIdleConns:       10,
		MaxConnAge:         0,
		PoolTimeout:        4000000000,
		IdleTimeout:        5000000000,
		IdleCheckFrequency: 60000000000,
		//   TLSConfig: (*tls.Config)(nil)
		//
		Dialer: func() (net.Conn, error) {
			netDialer := &net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Minute,
			}
			return netDialer.Dial("tcp", redisAddr)
		},

		// Hook
		OnConnect: func(conn *redis.Conn) error {
			fmt.Printf("Conn=%v\n", conn)
			return nil
		},
	}

	fmt.Printf("===== Liruhui add ===== redis.Options %#v\n", options)

	gRedis = redis.NewClient(&options)
	// defer gRedis.Close()

	// fmt.Printf("===== Liruhui add ===== gRedis %#v\n", gRedis)
	// 检查是否成功连接到了 redis 服务器
	pong, err := gRedis.Ping().Result()
	fmt.Println("===== Liruhui add ===== pong", pong, "err", err)
	// 设置一个键值对
	// err = gRedis.Set("province1", "hebei", 0).Err()
	if err != nil {
		fmt.Println("无法设置键值对:", err)
	} else {
		LoadCustInfoFile(CustInfoFilePath)
	}

	// handleRedisChannel()
	go handleRedisChannel()
}

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
	if len(custid) != 0 {
		fmt.Println("request custid:", custid)
	}
	if len(token) != 0 {
		fmt.Println("request token:", token)
	}

	if ok := blacklist1.SearchReqLimitByToken(token); ok {
		fmt.Println("token", token, "in ReqLimit")
	} else if ok := blacklist1.SearchReqLimitByCustId(custid, token); ok {
		fmt.Println("custId", custid, "in ReqLimit")
		// blacklist1.updateCacheToken(custid)
	} else {
		if len(token) != 0 || len(custid) != 0 {
			fmt.Println("custId", custid, "and token", token, "don't in ReqLimit")
		}
	}

	if print == "1" {
		blacklist1.PrintCacheReqLimit()
	}

	if reload == "1" {
		blacklist1.LoadConfigReqLimitFile(blacklist1.ReqLimitFilePath)
	}

	if clear == "1" {
		blacklist1.CacheCustId.Clear()
		blacklist1.CacheToken.Clear()
	}
}

func handleRequest1(c *gin.Context) {
	fmt.Println("hello")
}

func handleRequestBody(c *gin.Context) {
	// 读取请求体
	// body := make([]byte, c.Request.ContentLength)
	// _, err := c.Request.Body.Read(body)
	// defer c.Request.Body.Close()

	// 获取请求的Body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// blacklist1.CheckErr(err)

	str := string(body)
	fmt.Println("request body:", str)

	// 重置请求主体流的位置
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
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
	// custid := queryParams.Get("custid")
	token := queryParams.Get("token")
	print := queryParams.Get("print")
	reload := queryParams.Get("reload")
	clear := queryParams.Get("clear")

	// 获取客户信息
	if len(token) != 0 {
		custinfo := getCustInfo(token)
		custid := custinfo.Data.Data.CustID
		fmt.Println("request token:", token, "custid", custid)

		if ok := blacklist1.SearchReqLimitByToken(token); ok {
			fmt.Println("token", token, "in ReqLimit")
		} else if ok := blacklist1.SearchReqLimitByCustId(custid, token); ok {
			fmt.Println("custId", custid, "in ReqLimit")
			// blacklist1.updateCacheToken(custid)
		} else {
			if len(token) != 0 || len(custid) != 0 {
				fmt.Println("custId", custid, "and token", token, "don't in ReqLimit")
			}
		}
	}

	if print == "1" {
		blacklist1.PrintCacheReqLimit()
	}

	if reload == "1" {
		blacklist1.LoadConfigReqLimitFile(blacklist1.ReqLimitFilePath)
	}

	if clear == "1" {
		blacklist1.CacheCustId.Clear()
		blacklist1.CacheToken.Clear()
	}

	// for i := 0; i < 2; i++ {
	// 	handleRequestBody(c)
	// 	time.Sleep(2 * time.Second)
	// }
}

// func sendEmail() {
// 	fmt.Println("send email")
// }

// // cron
// func handleCron() {
// 	// 创建一个 Cron 对象
// 	cron := cron.New()

// 	// 设置每天 1:05 执行的任务
// 	job := cron.AddJob("5 1 * * *", sendEmail)

// 	// 启动 Cron 对象，开始监听定时任务
// 	cron.Start()

// 	// 等待一段时间，以便任务有机会执行
// 	time.Sleep(time.Hour * 24)

// 	// 停止 Cron 对象
// 	cron.Stop()

// }

func handleCron() {
	fmt.Println("create a timer")

	currentDate := time.Now()
	formattedDate := currentDate.Format("2006-01-02-15:04:05")
	fmt.Println(formattedDate)

	// 计算下一个20:05:01的时间
	nextTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), time.Now().Hour(), time.Now().Minute(), time.Now().Second()+10, 0, time.Local)
	duration := nextTime.Sub(time.Now())

	// 创建一个定时器
	timer := time.NewTimer(duration)

	fmt.Println("duration", duration)

	// 启动一个线程执行定时任务
	go func() {
		<-timer.C // 等待定时器触发的时间
		// 执行你的任务代码
		fmt.Println(time.Now(), "定时任务执行了！")
		timer.Stop()
		handleCron()
	}()
}

func logMongo() {
	if gCollection != nil {
		msg := logMsg{user: "test", age: 17}
		insert := bson.M{"name": msg.user, "age": msg.age}
		fmt.Println("insert:", insert)

		fmt.Println("atomic.LoadPointer(&p)", (*mongo.Collection)(atomic.LoadPointer(&UtilsMongoCol)))

		_, err := (*mongo.Collection)(atomic.LoadPointer(&UtilsMongoCol)).InsertOne(context.Background(), insert)
		// _, err := gCollection.InsertOne(context.Background(), insert)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func handleMongo() {
	Ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://admin:123456@192.168.225.128:27017/")
	fmt.Println("clientOptions", clientOptions)
	gMongo, err := mongo.Connect(Ctx, clientOptions)
	if err != nil {
		fmt.Printf("Failed to connect mongo with error : %v\n", err.Error())
	} else {
		err = gMongo.Ping(Ctx, nil)
		if err != nil {
			fmt.Printf("%s", err.Error())
		} else {
			// 集合格式：docker-compose.yml中day+日期+LOG_MONGO_COLLECTION
			// 示例：day20231112tracelog
			currentDate := time.Now()
			formattedDate := currentDate.Format("20060102")
			collectionName := strings.Join([]string{"day", formattedDate, "tracelog"}, "")
			gCollection = gMongo.Database("tms-api-gw-jh").Collection(collectionName)
			// utils.UtilsMongoCol = gCollection

			// 原子操作指针
			fmt.Println("gCollection", gCollection)
			atomic.StorePointer(&UtilsMongoCol, unsafe.Pointer(gCollection))
			fmt.Println("UtilsMongoCol", UtilsMongoCol)
			// fmt.Printf("UtilsMongoCol:%#v\n", UtilsMongoCol)

			fmt.Println("atomic.LoadPointer(&p)", atomic.LoadPointer(&UtilsMongoCol))

			logMongo()
		}
	}
}

func main() {
	// 黑名单初始化
	// blacklist1.Initblacklist()

	currentDate := time.Now()
	formattedDate := currentDate.Format("20060102150405")
	fmt.Println("formattedDate:", formattedDate)

	// 连接reids
	HandleRedis()

	// 定时任务
	// handleCron()

	// 连接mongo
	// fmt.Println("handle mongo")
	// handleMongo()

	blacklist1.InitReqLimit(gRedis)

	// 创建一个gin的路由器实例
	router := gin.Default()

	// 使用Any方法定义一个处理函数，处理所有类型的HTTP请求
	// router.GET("/blacklist", handleRequest)
	router.Any("/blacklist", handleRequest)
	// router.GET("/proxyPath/blacklist", handleRequest)
	// router.Any("/*proxyPath", handleRequest1)

	// 启动服务器并监听端口
	router.Run(":8080")

	// http.HandleFunc("/blacklist", handleCustId)

	// 启动HTTP服务器
	// err := http.ListenAndServe(":8080", nil)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	if gRedis != nil {
		fmt.Println("--------------- redis close ----------------")
		gRedis.Close()
	}
}
