package approuter

import (
	"api-gw/blacklist"
	"api-gw/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	logg   *logrus.Logger
	poolCh chan bool
	//Ipregex     *regexp.Regexp
	GlobalRules map[string]Rule
	gRedis      *redis.Client
	Ctx         context.Context
	gMongo      *mongo.Client
	gCollection *mongo.Collection
	//ClientMetrics map[string]*utils.ClientData
	//GlobalMetrics *utils.BaseData
	gBackupRedis *redis.Client
	gRedisMain   int
)

// metrics
var GwMetrics = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "api_gw_access_total",
	Help: "api_gw_access_total 当前新增访问量",
}, []string{"type"})

var GwClientMetrics = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "api_gw_client_access_total",
	Help: "api_gw_client_access_total 当前client新增访问量",
}, []string{"type", "client"})

type MetricsResponse struct {
	Code    string          `json:"code"`
	Message string          `json:"msg"`
	Data    *utils.BaseData `json:"data"`
}

type ProxyURL struct {
	Url   string `json:"url"`
	Label string `json:"label"`
}

// TODO target
type Rule struct {
	Targets           []ProxyURL `json:"target"`
	Auth              []string   `json:"auth"`
	TransformRequest  []string   `json:"transformRequest"`
	TransformResponse []string   `json:"transformResponse"`
	Trace             []string   `json:"trace"`
	Quota             []string   `json:"quota"`
	Timeout           int        `json:"timeout"`
	ProxyTimeout      int        `json:"proxyTimeout"`
}

// Init process
func Init(cfg utils.GlobalCfg, wg *sync.WaitGroup, sigc chan bool) error {
	//func Init(cfg utils.GlobalCfg) error {
	logg = logrus.New()
	logg.Formatter = &logrus.TextFormatter{
		TimestampFormat: "2006-01-02T15:04:05.999999999Z07:00",
		FullTimestamp:   true,
		DisableSorting:  true,
	}
	logFileName := path.Join(cfg.LogFilepath, cfg.LogFilename)
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Failed to open [%v] with error:[%v]\n", logFileName, err.Error())
		print("Failed to open file with error:" + err.Error())
		return err
	}

	logg.SetOutput(logFile)
	logg.SetLevel(logrus.Level(cfg.LogLevel))

	// max conn
	poolCh = make(chan bool, cfg.RestMaxConn)

	// TODO Mongo
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if cfg.MongoEnable {
		Ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
		clientOptions := options.Client().ApplyURI(cfg.MongoUrl)
		gMongo, err = mongo.Connect(Ctx, clientOptions)
		if err != nil {
			fmt.Printf("Failed to connect mongo with error : %v\n", err.Error())
			logg.Error(fmt.Sprintf("Failed to connect mongo with error :%v\n", err.Error()))
			//return err
		} else {
			err = gMongo.Ping(Ctx, nil)
			if err != nil {
				fmt.Printf("%s", err.Error())
				logg.Error(err.Error())
			} else {
				gCollection = gMongo.Database(cfg.MongoDB).Collection(cfg.MongoCollection)
			}
		}

		utils.LogMsgQueue = make(chan *utils.TraceLog, 10000)
		// Create mongo worker
		for i := 0; i < cfg.MongoWorkerNum; i++ {
			wg.Add(1)
			go utils.MongoWorker(gCollection, wg, sigc)
		}
	}

	// TODO Redis TODO
	redisAddr := cfg.RedisHost + ":" + cfg.RedisPort
	redisBackupAddr := cfg.RedisBackupHost + ":" + cfg.RedisBackupPort
	gRedisMain = cfg.RedisMainFlag

	gRedis = redis.NewClient(&redis.Options{
		// 连接信息
		Network:  "tcp",
		Addr:     redisAddr,
		Password: cfg.RedisPwd,
		DB:       0,

		// Pool
		PoolSize:     15,
		MinIdleConns: 10,

		// Timeout
		DialTimeout:  5 * time.Second,
		ReadTimeout:  time.Duration(int64(cfg.RedisReadTimeout)) * time.Millisecond,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,

		// Idle Time setting
		IdleCheckFrequency: 60 * time.Second,
		IdleTimeout:        5 * time.Second,
		MaxConnAge:         0 * time.Second,

		//
		MaxRetries:      0,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Microsecond,

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
	})

	gBackupRedis = redis.NewClient(&redis.Options{
		// 连接信息
		Network:  "tcp",
		Addr:     redisBackupAddr,
		Password: cfg.RedisBackupPwd,
		DB:       0,

		// Pool
		PoolSize:     15,
		MinIdleConns: 10,

		// Timeout
		DialTimeout:  5 * time.Second,
		ReadTimeout:  time.Duration(int64(cfg.RedisReadTimeout)) * time.Millisecond,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,

		// Idle Time setting
		IdleCheckFrequency: 60 * time.Second,
		IdleTimeout:        5 * time.Second,
		MaxConnAge:         0 * time.Second,

		//
		MaxRetries:      0,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Microsecond,

		//
		Dialer: func() (net.Conn, error) {
			netDialer := &net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Minute,
			}
			return netDialer.Dial("tcp", redisBackupAddr)
		},

		// Hook
		OnConnect: func(conn *redis.Conn) error {
			fmt.Printf("Conn=%v\n", conn)
			return nil
		},
	})

	// Rules load
	GlobalRules = make(map[string]Rule)
	if err := json.Unmarshal([]byte(utils.GCfg.ProxyRules), &GlobalRules); err != nil {
		fmt.Printf("Failed to marshal the rules with error: %s\n", err.Error())
		logg.Error(fmt.Sprintf("Failed to marshal the rules with error: [%s]\n", err.Error()))
	}

	//GlobalMetrics = &utils.BaseData{
	//	SlowNum:        0,
	//	TotalNum:       0,
	//	SendTotalNum:   0,
	//	SendFailNum:    0,
	//	SendSuccessNum: 0,
	//	SendErrorNum:   0,
	//}
	//ClientMetrics = make(map[string]*utils.ClientData)

	// 初始化客户ID黑名单
	blacklistConfig := blacklist.InitConfigBlacklist{BlacklistFilePath: cfg.BlacklistConfigName, CustIdMaxSize: cfg.BlacklistConfigCustidMax, TokenExpireTime: cfg.BlacklistConfigTokenExpiration, TokenMaxSize: cfg.BlacklistConfigTokenMax}
	blacklist.Initblacklist(blacklistConfig)

	return nil
}

// Uninit
func Uninit() {
	// Redis
	if gRedis != nil {
		gRedis.Close()
	}

	if gBackupRedis != nil {
		gBackupRedis.Close()
	}

	// Mongo
	if gMongo != nil {
		gMongo.Disconnect(Ctx)
	}
}

// Get the Rules from Request
func GetRules(c *gin.Context) (*Rule, string, string, string) {
	var targetRule *Rule
	path := c.Request.URL.Path

	urlPrefix := ""
	originUrl := path
	newReqUrl := ""

	// Fix
	path = strings.TrimRight(path, "/")
	fmt.Printf("The url-path: %s\n", c.Request.URL.Path)
	logg.Debug(fmt.Sprintf("The url-path: %s\n", c.Request.URL.Path))

	// 2) Foreach keys in map to match the url, if ok, return
	var rule Rule
	findFlag := false
	rule, ok := GlobalRules[path]
	if !ok {
		//fmt.Printf("Failed to get the rule in map by key[%v]\n", path)
		for key, val := range GlobalRules {
			//fmt.Printf("key in map: [%v]\n", key)
			logg.Debug(fmt.Sprintf("Key in map: [%v]\n", key))
			if strings.HasPrefix(path, key) {
				urlPrefix = key
				targetRule = &val
				newReqUrl = strings.TrimPrefix(path, key)
				//TODO combine the path
				findFlag = true
				break
			}
		}
	} else {
		findFlag = true
		targetRule = &rule
	}
	if findFlag == true {
		fmt.Printf("Got the rule: [%v]\n", targetRule)
		logg.Debug(fmt.Sprintf("Got the rule: [%v]\n", targetRule))
	} else {
		fmt.Printf("Got no rule by path [%v]\n", path)
		logg.Error(fmt.Sprintf("Got no rule by path [%v]\n", path))
		return nil, path, path, ""
	}
	// 2-both) cache the keys to do the next step
	// 3) if not, hasPrefix to search/match the target
	return targetRule, urlPrefix, originUrl, newReqUrl
}

// Main block before send
func PreHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 处理黑名单接口
		blacklist.BlacklistHandleRequest(c)

		//nowTime := time.Now().UnixNano()
		nowTime := time.Now()
		// Create the log handle
		traceLogObj := &utils.TraceLog{}
		traceLogObj.OrgRequest = fmt.Sprintf("%v", *c.Request)

		//==============================================
		// Set
		//   x-request-id: uuid
		//   x-request-at: timestring
		//   x-request-ip: ip
		// for Request
		//==============================================
		strUUID := uuid.New().String()
		xrequestid := c.GetHeader("x-request-id")
		if xrequestid != "" {
			strUUID = xrequestid
			c.Request.Header.Del("x-request-id")
		}
		// !!!! Add will call CanonicalMIMEHeaderKey to make header Uper-Case
		c.Request.Header["x-request-id"] = []string{strUUID}
		c.Set("UUID", strUUID)
		// DONE (1)
		traceLogObj.RequestId = strUUID
		timeAt := time.Now().String()

		//c.Request.Header.Add("x-request-at", timeAt)
		c.Request.Header["x-request-at"] = []string{timeAt}
		traceLogObj.RequestAt = timeAt

		// ClientIP implements one best effort algorithm to return the real client IP.
		// It calls c.RemoteIP() under the hood, to check if the remote IP is a trusted proxy or not.
		// If it is it will then try to parse the headers defined in Engine.RemoteIPHeaders (defaulting to [X-Forwarded-For, X-Real-Ip]).
		// If the headers are not syntactically valid OR the remote IP does not correspond to a trusted proxy,
		// the remote IP (coming from Request.RemoteAddr) is returned.
		clientIp := c.ClientIP()
		//c.Request.Header.Add("x-request-ip", clientIp)
		c.Request.Header["x-request-ip"] = []string{clientIp}
		traceLogObj.RequestIp = clientIp

		// Update the counter for metrics
		// Atomic TODO
		//atomic.AddInt64(&GlobalMetrics.TotalNum, 1)
		//GlobalMetrics.TotalNum += 1

		// Get the rules
		targetRule, urlPrefix, originUrl, newReqUrl := GetRules(c)

		fmt.Printf("Got the rule: \ntargetRule:[%v],\nurlPrefix: [%v],\noriginUrl: [%v],\nnewReqUrl:[%v]\n", targetRule, urlPrefix, originUrl, newReqUrl)
		//logg.Debug(fmt.Sprintf("Got the rule: \ntargetRule:[%v],\nurlPrefix: [%v],\noriginUrl: [%v],\nnewReqUrl:[%v]\n", targetRule, urlPrefix, originUrl, newReqUrl))
		traceLogObj.Rule = fmt.Sprintf("%v", *targetRule)
		if targetRule == nil {
			// No Rules
			c.JSON(404, gin.H{
				"Code": "10001",
				"Msg":  "Found no target Rule",
			})
			// Before Abort, Logg
			traceLogObj.ErrorMsg = "Found no target Rule"
			logg.Error(utils.GetTraceLog(traceLogObj))
			c.Abort()
			return
		} else {
			//
			c.Set("targetRule", targetRule)
			c.Set("urlPrefix", urlPrefix)
			c.Set("originUrl", originUrl)
			c.Set("newReqUrl", newReqUrl)
		}

		//log
		if logg != nil {
			c.Set("logHandle", logg)
		}
		c.Set("traceLogObj", traceLogObj)
		if utils.GCfg.AuthEnable == true {
			c.Set("redisHandle", gRedis)
			c.Set("redisBackupHandle", gBackupRedis)
			c.Set("redisMainFlag", &gRedisMain)
		}
		// Metrics
		//c.Set("gMetrics", GlobalMetrics)
		//c.Set("cMetrics", ClientMetrics)

		// promethus metrics
		c.Set("gwm", GwMetrics)
		c.Set("gwcm", GwClientMetrics)
		GwMetrics.WithLabelValues("total").Inc()

		c.Next()

		costTime := time.Since(nowTime)
		//costTime := time.Now().UnixNano() - nowTime
		//traceLogObj.TotalEscaped = fmt.Sprintf("%v", costTime/1000000.000000)
		traceLogObj.TotalEscaped = fmt.Sprintf("%v", utils.GetMillseconds(costTime))
		//update response
		traceLogObj.NewResponse = fmt.Sprintf("%v", c.Request.Response)
		utils.LoggerIt(c, "info", "")
	}
}

// proxy
func ProxyHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceLog, ok := c.Get("traceLogObj")
		var traceLogObj *utils.TraceLog
		if !ok {
			logg.Error("Found no traceLogObj in modifyResponse")
		} else {
			traceLogObj = traceLog.(*utils.TraceLog)
		}

		//nowTime := time.Now().UnixNano()
		nowTime := time.Now()
		//var costTime int64
		var costTime time.Duration
		var custMetrics *utils.ClientData
		var custId interface{}
		//var appId string
		// Auth Check
		target, ok := c.Get("targetRule")
		targetURL := ""
		if !ok {
			//TODO No target URL
			c.Abort()
			// Metrics
			GwMetrics.WithLabelValues("sendFail").Inc()
			//atomic.AddInt64(&GlobalMetrics.SendFailNum, 1)
			logg.Error(fmt.Sprintf("The request [%v] found no target rule\n", c.Request))
			c.JSON(http.StatusBadRequest, gin.H{
				"code": "10003",
				"msg":  "Found no target Rule",
			})
			return
		} else {

			//rawAppId, ok := c.Get("AppId")
			//if ok {
			//	appId = rawAppId.(string)
			//}
			// CustID & CustLable to found the target
			// Docker Mode: ENV from AuthClientIDField: clientId
			custId, ok = c.Get("CustId")
			if !ok {
				fmt.Printf("Failed to fetch the custId\n")
				// TODO
				c.Abort()
				GwMetrics.WithLabelValues("sendFail").Inc()
				//if appId != "" {
				//	GwClientMetrics.WithLabelValues("clientSendFail", appId).Inc()
				//}
				//atomic.AddInt64(&GlobalMetrics.SendFailNum, 1)
				logg.Error(fmt.Sprintf("The request [%v] has no custId\n", c.Request))
				c.JSON(http.StatusBadRequest, gin.H{
					"code": "10001",
					"msg":  "No CustID",
				})
				return
			} else {
				if custId != nil && custId.(string) != "" {
					//c.Request.Header.Add("x-request-custid", custId.(string))
					c.Request.Header["x-request-custid"] = []string{custId.(string)}
					if traceLogObj != nil {
						traceLogObj.RequestCustId = custId.(string)
					}
				}
			}

			// Metrics
			//GwClientMetrics.WithLabelValues("clientSendTotal", appId).Inc()
			GwClientMetrics.WithLabelValues("clientSendTotal", custId.(string)).Inc()
			//custMetrics = ClientMetrics[custId.(string)]
			//if custMetrics == nil {
			//	// Add new one
			//	clientData := new(utils.ClientData)
			//	clientData.ClientInfo = c.GetHeader("x-request-id")
			//	atomic.AddInt64(&clientData.Counter.TotalNum, 1)
			//	ClientMetrics[custId.(string)] = clientData
			//} else {
			//	// Update
			//	atomic.AddInt64(&custMetrics.Counter.TotalNum, 1)
			//}

			// CustLabel
			custLabel, ok := c.Get("CustLabel")
			if !ok {
				fmt.Printf("Failed to get custLabel to find the target URL, so using the first one\n")
				logg.Debug(fmt.Sprintf("Request [%v]: failed to get custLabel to find the target URL, so using the first one\n", c.Request))
				targetURL = target.(*Rule).Targets[0].Url
				// TODO Using the default value
			} else {
				for _, t := range target.(*Rule).Targets {
					if t.Label == custLabel.(string) {
						// Find the target
						targetURL = t.Url
						break
					}
				}
			}
		}

		// Parse URL
		parsedURL, err := url.Parse(targetURL)
		if err != nil {
			fmt.Printf("Failed to parse URL with error :[%v]\n", err.Error())
			logg.Error(fmt.Sprintf("Request [%v]: failed to parse the URL [%v] with error [%v]\n", c.Request, targetURL, err.Error()))
			c.Abort()
			//atomic.AddInt64(&GlobalMetrics.SendFailNum, 1)
			GwMetrics.WithLabelValues("sendFail").Inc()
			GwClientMetrics.WithLabelValues("clientSendFail", custId.(string)).Inc()
			if custMetrics != nil {
				atomic.AddInt64(&custMetrics.Counter.SendFailNum, 1)
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"code": "10002",
				"msg":  err.Error(),
			})
			return
		}

		fmt.Printf("Request [%v]: parsed Scheme [%v], Host [%v], Path [%v] by targetURL [%v]\n", c.Request, parsedURL.Scheme, parsedURL.Host, parsedURL.Path, targetURL)
		logg.Debug(fmt.Sprintf("Request [%v]: parsed Scheme [%v], Host [%v], Path [%v] by targetURL [%v]\n", c.Request, parsedURL.Scheme, parsedURL.Host, parsedURL.Path, targetURL))

		path, ok := c.Get("newReqUrl")
		if !ok {
			fmt.Printf("Failed to get newReqUrl")
			logg.Warn("Failed to get newReqUrl")
		}

		newPath := parsedURL.Path
		if path != nil && path.(string) != "" {
			newPath = utils.MakePath(newPath, path.(string))
		}

		// TODO Just be ready
		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				//targetQuery := c.Request.URL.RawQuery
				req.URL.Scheme = parsedURL.Scheme
				req.URL.Host = parsedURL.Host
				req.URL.Path = newPath
				req.Header = c.Request.Header
				req.Host = c.Request.Host
				// Query
				//if targetQuery == "" || req.URL.RawQuery == "" {
				//	req.URL.RawQuery = targetQuery + req.URL.RawQuery
				//} else {
				//	req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
				//}
			},
		}

		// Response
		proxy.ModifyResponse = func(resp *http.Response) error {
			//atomic.AddInt64(&GlobalMetrics.SendSuccessNum, 1)
			GwMetrics.WithLabelValues("sendSuccess").Inc()
			//GwClientMetrics.WithLabelValues("clientSendSuccess", appId).Inc()
			GwClientMetrics.WithLabelValues("clientSendSuccess", custId.(string)).Inc()
			//if custMetrics != nil {
			//	atomic.AddInt64(&custMetrics.Counter.SendSuccessNum, 1)
			//}
			costTime = time.Since(nowTime)
			//costTime = time.Now().UnixNano() - nowTime
			// update OrgResponse
			if traceLogObj != nil {
				traceLogObj.OrgResponse = fmt.Sprintf("%v", *resp)
				traceLogObj.NewRequest = fmt.Sprintf("%v", *resp.Request)
				traceLogObj.NewResponseCode = fmt.Sprintf("%v", resp.StatusCode)
			}

			fmt.Printf("resp: [%v]\n", resp)
			fmt.Printf("resp's new-request: [%v]\n", resp.Request)
			if resp.StatusCode != 200 {
				oldPayload, err := io.ReadAll(resp.Body)
				if err != nil {
					return err
				}
				newPadload := []byte("ProxyStatusCode error:" + string(oldPayload))
				resp.Body = io.NopCloser(bytes.NewBuffer(newPadload))
				resp.ContentLength = int64(len(newPadload))
				resp.Header.Set("Content-Length", strconv.FormatInt(int64(len(newPadload)), 10))
			}
			return nil
		}

		// Error && ErrorHandler
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			//costTime = time.Now().UnixNano() - nowTime
			costTime = time.Since(nowTime)
			if err != nil {
				//atomic.AddInt64(&GlobalMetrics.SendErrorNum, 1)
				GwMetrics.WithLabelValues("sendError").Inc()
				//GwClientMetrics.WithLabelValues("clientSendError", appId).Inc()
				GwClientMetrics.WithLabelValues("clientSendError", custId.(string)).Inc()
				//if custMetrics != nil {
				//	atomic.AddInt64(&custMetrics.Counter.SendErrorNum, 1)
				//}
				fmt.Printf("Error: ErrorHandler catched error: [%v]\n", err.Error())
				fmt.Printf("Content-Length: [%s]\n", r.Header.Get("Content-Length"))
				logg.Error(fmt.Sprintf("ErrorHandler catched error: [%v]\n", err.Error()))

				w.WriteHeader(http.StatusBadGateway)
				_, _ = fmt.Fprintf(w, err.Error())

			}
		}

		// TODO
		proxy.ServeHTTP(c.Writer, c.Request)
		c.Next()
		if traceLogObj != nil {
			//traceLogObj.BindEscaped = fmt.Sprintf("%v", costTime/1000000.000000)
			traceLogObj.BindEscaped = fmt.Sprintf("%v", utils.GetMillseconds(costTime))
		}
	}
}
