package utils

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/Unknwon/goconfig"
	"github.com/sirupsen/logrus"
)

type GlobalCfg struct {
	//log
	LogFilepath string
	LogLevel    int
	LogFilename string
	//
	RestPort       string
	CtrlPort       string
	RestMaxConn    int
	LogQueueLength int
	LogQueue       chan (string)
	MaxCpuNum      int
	ReadTimeout    int
	WriteTimeout   int
	IdleTimeout    int

	// Mongo
	MongoEnable     bool
	MongoUrl        string
	MongoDB         string
	MongoCollection string
	MongoWorkerNum  int

	// Redis
	RedisMainFlag    int
	RedisHost        string
	RedisPort        string
	RedisPwd         string
	RedisBackupHost  string
	RedisBackupPort  string
	RedisBackupPwd   string
	RedisReadTimeout int

	// Proxy
	ProxyKeepalive bool
	ProxyOptions   string
	ProxyRules     string

	// TransformRequest
	TReqEnable  bool
	TReqDefault string

	// TransformResponse
	TResEnable  bool
	TResDefault string

	// Trace
	TraceEnable  bool
	TraceDefault string

	// Quota
	QuotaEnable bool

	// Auth
	AuthEnable      bool
	AuthKeepalive   bool
	AuthOptions     string
	AuthDefault     string
	AuthHttpURL     string
	AuthTimeout     int
	AuthWhiteIPList string
	// ++++
	AuthCUSTIDField    string
	AuthCUSTLabelField string
	AuthCUSTIDPingan   string

	// PushMessage
	PMEnable       bool
	PMLogpath      string
	PMPrefix       string
	PMRedisHost    string
	PMRedisPort    string
	PMRedisPwd     string
	PMRedisChannel string

	// API
	ApiEnable bool
	ApiPort   string

	// Blacklist config
	BlacklistConfigName            string
	BlacklistConfigTokenMax        int
	BlacklistConfigCustidMax       int
	BlacklistConfigTokenExpiration int64
}

// Global Variable
var GCfg GlobalCfg

func ParseConfig(dockerStyle bool) error {
	if dockerStyle == false {
		cfg, err := goconfig.LoadConfigFile("config.ini")
		if err != nil {
			print("Failed to load config.ini")
			return err
		}

		//Item
		itemLog := "log"
		itemRest := "common"
		itemMongo := "mongo"
		itemRedis := "redis"
		itemProxy := "proxy"
		itemTReq := "transformRequest"
		itemTRes := "transformResponse"
		itemTrace := "trace"
		itemQuota := "quota"
		itemAuth := "auth"
		itemPushMessage := "pushmessage"
		itemApi := "api"
		itemBlacklist := "blacklist"

		// log
		logfilepath, _ := cfg.GetValue(itemLog, "logfilepath")
		GCfg.LogFilepath = logfilepath
		logfilename, _ := cfg.GetValue(itemLog, "logfilename")
		GCfg.LogFilename = logfilename
		loglevel, _ := cfg.GetValue(itemLog, "loglevel")
		switch loglevel {
		case "error":
			GCfg.LogLevel = int(logrus.ErrorLevel)
		case "warn", "warning":
			GCfg.LogLevel = int(logrus.WarnLevel)
		case "info":
			GCfg.LogLevel = int(logrus.InfoLevel)
		case "debug":
			GCfg.LogLevel = int(logrus.DebugLevel)
		default:
			GCfg.LogLevel = int(logrus.ErrorLevel)
		}
		// Rest
		restport, _ := cfg.GetValue(itemRest, "port")
		GCfg.RestPort = restport
		ctrlPort, _ := cfg.GetValue(itemRest, "ctrlPort")
		GCfg.CtrlPort = ctrlPort
		restMaxConn, _ := cfg.GetValue(itemRest, "maxConn")
		irestMaxConn, err := strconv.Atoi(restMaxConn)
		if err != nil {
			print(fmt.Sprintf("Invalid maxConn, check restMaxConn in config.ini:%v\n", restMaxConn))
			return errors.New("Invalid maxConn, non-integer")
		}
		GCfg.RestMaxConn = irestMaxConn
		lenLogQ, _ := cfg.GetValue(itemRest, "logQueueLength")
		ilenLogQ, err := strconv.Atoi(lenLogQ)
		if err != nil {
			print(fmt.Sprintf("Invalid logQueueLength, check it in config.ini\n"))
			return errors.New("Invalid logQueueLength, non-integer")
		}
		GCfg.LogQueueLength = ilenLogQ
		GCfg.LogQueue = make(chan string, GCfg.LogQueueLength)
		maxC, _ := cfg.GetValue(itemRest, "maxcpunum")
		iMaxC, err := strconv.Atoi(maxC)
		if err != nil {
			print(fmt.Sprintf("Invalid maxCpuNum, check maxcpunum in config.ini:%v\n", maxC))
			return errors.New("Invalid maxcpunum, non-integer")
		}
		GCfg.MaxCpuNum = iMaxC
		// ReadTimeout
		rto, _ := cfg.GetValue(itemRest, "readtimeout")
		irto, err := strconv.Atoi(rto)
		if err != nil {
			print(fmt.Sprintf("Invalid readtimeout, check readtimeout in config.ini:%v\n", rto))
			return errors.New("Invalid readtimeout, non-integer")
		}
		GCfg.ReadTimeout = irto
		// WriteTimeout
		wto, _ := cfg.GetValue(itemRest, "writetimeout")
		iwto, err := strconv.Atoi(wto)
		if err != nil {
			print(fmt.Sprintf("Invalid writetimeout, check writetimeout in config.ini:%v\n", wto))
			return errors.New("Invalid writetimeout, non-integer")
		}
		GCfg.WriteTimeout = iwto
		// IdleTimeout
		ito, _ := cfg.GetValue(itemRest, "idletimeout")
		iito, err := strconv.Atoi(ito)
		if err != nil {
			print(fmt.Sprintf("Invalid idletimeout, check idletimeout in config.ini:%v\n", ito))
			return errors.New("Invalid idletimeout, non-integer")
		}
		GCfg.IdleTimeout = iito

		// Mongo
		//MongoEnable     bool
		menable, _ := cfg.GetValue(itemMongo, "enable")
		bmenable, err := strconv.ParseBool(menable)
		if err != nil {
			print(fmt.Sprintf("Invalid mongo-enable, check enable in config.ini:%s\n", menable))
			return errors.New("Invalid mongo-enable, non-bool")
		}
		GCfg.MongoEnable = bmenable
		//MongoUrl
		mu, _ := cfg.GetValue(itemMongo, "url")
		GCfg.MongoUrl = mu
		//MongoDB         string
		mdb, _ := cfg.GetValue(itemMongo, "database")
		GCfg.MongoDB = mdb
		//MongoCollection string
		mc, _ := cfg.GetValue(itemMongo, "collection")
		GCfg.MongoCollection = mc
		//MongoWorkerNum  int
		mwn, _ := cfg.GetValue(itemMongo, "workernum")
		imwn, err := strconv.Atoi(mwn)
		if err != nil {
			print(fmt.Sprintf("Invalid mongo-workernum, check workernum in config.ini:%s\n", mwn))
			return errors.New("Invalid mongo-workernum, non-integer")
		}
		GCfg.MongoWorkerNum = imwn

		// Redis
		// RedisMainFlag
		rmainflag, _ := cfg.GetValue(itemRedis, "mainflag")
		irmainflag, err := strconv.Atoi(rmainflag)
		if err != nil {
			print(fmt.Sprintf("Invalid redis-mainflag, check mainflag in config.ini:%s\n", &rmainflag))
			return errors.New("Invalid redis-mainflag, non-bool")
		}
		GCfg.RedisMainFlag = irmainflag
		//RedisHost    string
		rhost, _ := cfg.GetValue(itemRedis, "mainhost")
		GCfg.RedisHost = rhost
		//RedisPort    string
		rport, _ := cfg.GetValue(itemRedis, "mainport")
		GCfg.RedisPort = rport
		//RedisPwd     string
		rpwd, _ := cfg.GetValue(itemRedis, "mainpassword")
		GCfg.RedisPwd = rpwd
		rbhost, _ := cfg.GetValue(itemRedis, "backuphost")
		GCfg.RedisBackupHost = rbhost
		rbport, _ := cfg.GetValue(itemRedis, "backupport")
		GCfg.RedisBackupPort = rbport
		rbpwd, _ := cfg.GetValue(itemRedis, "backuppassword")
		GCfg.RedisBackupPwd = rbpwd

		//RedisReadTimeout int
		rrt, _ := cfg.GetValue(itemRedis, "readtimeout")
		irrt, err := strconv.Atoi(rrt)
		if err != nil {
			print(fmt.Sprintf("Invalid readtimeout in config.ini:%v\n", rrt))
			return errors.New("Invalid readtimeout in config.ini, non-int")
		}
		GCfg.RedisReadTimeout = irrt

		// Proxy
		//ProxyKeepalive bool
		pkeepalive, _ := cfg.GetValue(itemProxy, "keepalive")
		bpkeepalive, err := strconv.ParseBool(pkeepalive)
		if err != nil {
			print(fmt.Sprintf("Invalid keepalive in config.ini:%v\n", pkeepalive))
			return errors.New("Invalid keepalive in config.ini, non-bool")
		}
		GCfg.ProxyKeepalive = bpkeepalive
		//ProxyOptions   string
		poptions, _ := cfg.GetValue(itemProxy, "agentOptions")
		GCfg.ProxyOptions = poptions
		//ProxyRules     string
		prules, _ := cfg.GetValue(itemProxy, "rules")
		GCfg.ProxyRules = prules

		// TransformRequest
		//TReqEnable  bool
		treqE, _ := cfg.GetValue(itemTReq, "enable")
		btreqE, err := strconv.ParseBool(treqE)
		if err != nil {
			print(fmt.Sprintf("Invalid transportRequest-enable in config.ini:%v\n", treqE))
			return errors.New("Invalid transportRequest-enable in config.ini")
		}
		GCfg.TReqEnable = btreqE
		//TReqDefault string
		treqDef, _ := cfg.GetValue(itemTReq, "default")
		GCfg.TReqDefault = treqDef

		// TransformResponse
		//TResEnable  bool
		tresE, _ := cfg.GetValue(itemTRes, "enable")
		btresE, err := strconv.ParseBool(tresE)
		if err != nil {
			print(fmt.Sprintf("Invalid transportResponse-enable in config.ini:%v\n", tresE))
			return errors.New("Invalid transportResponse-enable in config.ini")
		}
		GCfg.TResEnable = btresE
		//TResDefault string
		tresD, _ := cfg.GetValue(itemTRes, "default")
		GCfg.TResDefault = tresD

		// Trace
		//TraceEnable  bool
		trE, _ := cfg.GetValue(itemTrace, "enable")
		btrE, err := strconv.ParseBool(trE)
		if err != nil {
			print(fmt.Sprintf("Invalid trace-Enable in config.ini:%v\n", trE))
			return errors.New("Invalid trace-Enable in config.ini")
		}
		GCfg.TraceEnable = btrE
		//TraceDefault string
		trD, _ := cfg.GetValue(itemTrace, "default")
		GCfg.TraceDefault = trD

		// Quota
		//QuotaEnable bool
		qE, _ := cfg.GetValue(itemQuota, "enable")
		bqE, err := strconv.ParseBool(qE)
		if err != nil {
			print(fmt.Sprintf("Invalid quota-enable in config.ini:%v\n", qE))
			return errors.New("Invalid quota-enable in config.ini")
		}
		GCfg.QuotaEnable = bqE

		// Auth
		//AuthEnable    bool
		authE, _ := cfg.GetValue(itemAuth, "enable")
		bauthE, err := strconv.ParseBool(authE)
		if err != nil {
			print(fmt.Sprintf("Invalid auth-enable in config.ini:%v\n", authE))
			return errors.New("Invalid auth-enable in config.ini")
		}
		GCfg.AuthEnable = bauthE
		//AuthKeepalive bool
		authKeepalive, _ := cfg.GetValue(itemAuth, "keepalive")
		bauthKeepalive, err := strconv.ParseBool(authKeepalive)
		if err != nil {
			print(fmt.Sprintf("Invalid auth-keepalive in config.ini:%v\n", authKeepalive))
			return errors.New("Invalid auth-keepalive in config.ini")
		}
		GCfg.AuthKeepalive = bauthKeepalive
		//AuthOptions   string
		authOptions, _ := cfg.GetValue(itemAuth, "agentOptions")
		GCfg.AuthOptions = authOptions
		//AuthDefault   string
		authD, _ := cfg.GetValue(itemAuth, "default")
		GCfg.AuthDefault = authD
		authHttpURL, _ := cfg.GetValue(itemAuth, "httpURL")
		GCfg.AuthHttpURL = authHttpURL
		// AuthTimeout
		authTimeout, _ := cfg.GetValue(itemAuth, "timeout")
		iAuthTimeout, err := strconv.Atoi(authTimeout)
		if err != nil {
			print(fmt.Sprintf("Invalid AuthTimeout in config.ini:%v\n", authTimeout))
			return errors.New("Invalid AuthTimeout in config.ini")
		}
		GCfg.AuthTimeout = iAuthTimeout
		// AuthWhiteIPList
		authWhiteIPList, _ := cfg.GetValue(itemAuth, "whiteIPList")
		GCfg.AuthWhiteIPList = authWhiteIPList
		//AuthCUSTIDField    string
		custID, _ := cfg.GetValue(itemAuth, "custID")
		GCfg.AuthCUSTIDField = custID
		//AuthCustLabelField string
		custLabel, _ := cfg.GetValue(itemAuth, "custLabel")
		GCfg.AuthCUSTLabelField = custLabel
		//AuthCUSTIDPingan string
		custIDPingan, _ := cfg.GetValue(itemAuth, "custIDPingan")
		GCfg.AuthCUSTIDPingan = custIDPingan

		// PushMessage
		//PMEnable       bool
		pmE, _ := cfg.GetValue(itemPushMessage, "enable")
		bpmE, err := strconv.ParseBool(pmE)
		if err != nil {
			print(fmt.Sprintf("Invalid pushmessage-enable in config.ini:%v\n", pmE))
			return errors.New("Invalid pushmessage-enable in config.ini")
		}
		GCfg.PMEnable = bpmE
		//PMLogpath      string
		pmlogpath, _ := cfg.GetValue(itemPushMessage, "logpath")
		GCfg.PMLogpath = pmlogpath
		//PMRedisHost    string
		pmredishost, _ := cfg.GetValue(itemPushMessage, "redisHost")
		GCfg.PMRedisHost = pmredishost
		//PMRedisPort    string
		pmredisport, _ := cfg.GetValue(itemPushMessage, "redisPort")
		GCfg.PMRedisPort = pmredisport
		//PMRedisPwd     string
		pmredispwd, _ := cfg.GetValue(itemPushMessage, "redisPwd")
		GCfg.PMRedisPwd = pmredispwd

		// API
		//ApiEnable bool
		apiE, _ := cfg.GetValue(itemApi, "enable")
		bapiE, err := strconv.ParseBool(apiE)
		if err != nil {
			print(fmt.Sprintf("Invalid api-enable in config.ini:%v\n", apiE))
			return errors.New("Invalid api-enable in config.ini")
		}
		GCfg.ApiEnable = bapiE
		//ApiPort   string
		apiport, _ := cfg.GetValue(itemApi, "port")
		GCfg.ApiPort = apiport

		// 黑名单配置解析
		blacklistConfigName, _ := cfg.GetValue(itemBlacklist, "name")
		GCfg.BlacklistConfigName = blacklistConfigName
		iBlacklistConfigTokenMax, _ := cfg.GetValue(itemBlacklist, "tokenMax")
		blacklistConfigTokenMax, err := strconv.Atoi(iBlacklistConfigTokenMax)
		if err != nil {
			fmt.Println("Invalid BLACKLIST_CONFIG_CACHE_TOKEN_MAX, check it in docker-compose.yml")
			return errors.New("解析config.ini中黑名单配置本地缓存TOKEN最大值错误")
		}
		GCfg.BlacklistConfigTokenMax = blacklistConfigTokenMax
		iBlacklistConfigCustidMax, _ := cfg.GetValue(itemBlacklist, "custIdMax")
		blacklistConfigCustidMax, err := strconv.Atoi(iBlacklistConfigCustidMax)
		if err != nil {
			fmt.Println("Invalid BLACKLIST_CONFIG_CACHE_CUSTID_MAX, check it in docker-compose.yml")
			return errors.New("解析config.ini中黑名配置单本地缓存CUSTID最大值错误")
		}
		GCfg.BlacklistConfigCustidMax = blacklistConfigCustidMax
		iBlacklistConfigExpiration, _ := cfg.GetValue(itemBlacklist, "expiration")
		blacklistConfigExpiration, err := strconv.ParseInt(iBlacklistConfigExpiration, 10, 64)
		if err != nil {
			fmt.Println("Invalid BLACKLIST_CONFIG_TOKEN_EXPIRATION_TIME, check it in docker-compose.yml")
			return errors.New("解析config.ini中黑名单配置TOEKN本地缓存过期时间错误")
		}
		GCfg.BlacklistConfigTokenExpiration = blacklistConfigExpiration
	} else {
		/*
		 *Load config from docker env
		 */

		//log
		//LogFilepath string
		logFP := os.Getenv("LOG_FILEPATH")
		GCfg.LogFilepath = logFP
		//LogLevel    int
		logLvl := os.Getenv("LOG_LEVEL")
		switch logLvl {
		case "error":
			GCfg.LogLevel = int(logrus.ErrorLevel)
		case "warn", "warning":
			GCfg.LogLevel = int(logrus.WarnLevel)
		case "info":
			GCfg.LogLevel = int(logrus.InfoLevel)
		case "debug":
			GCfg.LogLevel = int(logrus.DebugLevel)
		default:
			GCfg.LogLevel = int(logrus.ErrorLevel)
		}
		//LogFilename string
		logFN := os.Getenv("LOG_FILENAME")
		GCfg.LogFilename = logFN

		////
		//RestPort       string
		rp := os.Getenv("APP_PORT")
		GCfg.RestPort = rp
		ctrlP := os.Getenv("APP_CTRLPORT")
		GCfg.CtrlPort = ctrlP
		//RestMaxConn    int
		restMC := os.Getenv("APP_MAXCONN")
		irestMC, err := strconv.Atoi(restMC)
		if err != nil {
			print(fmt.Sprintf("Invalid APP_MAXCONN, check it in docker-compose.yml\n"))
			return errors.New("Invalid APP_MAXCONN, non-integer")
		}
		GCfg.RestMaxConn = irestMC
		//LogQueueLength int
		//LogQueue       chan (string)
		// MaxCpuNum
		maxCpuNum := os.Getenv("APP_MAXCPUNUM")
		iMaxCpuNum, err := strconv.Atoi(maxCpuNum)
		if err != nil {
			print(fmt.Sprintf("Invalid APP_MAXCPUNUM, check it in docker-compose.yml\n"))
			return errors.New("Invalid APP_MAXCPUNUM, non-integer")
		}
		GCfg.MaxCpuNum = iMaxCpuNum
		// ReadTimeout
		readTimeout := os.Getenv("APP_READ_TIMEOUT")
		ireadTimeout, err := strconv.Atoi(readTimeout)
		if err != nil {
			print(fmt.Sprintf("Invalid APP_READ_TIMEOUT, check it in docker-compose.yml\n"))
			return errors.New("Invalid APP_READ_TIMEOUT")
		}
		GCfg.ReadTimeout = ireadTimeout
		// WriteTimeout
		writeTimeout := os.Getenv("APP_WRITE_TIMEOUT")
		iwriteTimeout, err := strconv.Atoi(writeTimeout)
		if err != nil {
			print(fmt.Sprintf("Invalid APP_WRITE_TIMEOUT, check it in docker-compose.yml\n"))
			return errors.New("Invalid APP_WRITE_TIMEOUT")
		}
		GCfg.WriteTimeout = iwriteTimeout
		// IdleTimeout
		idleTimeout := os.Getenv("APP_IDLE_TIMEOUT")
		iidleTimeout, err := strconv.Atoi(idleTimeout)
		if err != nil {
			print(fmt.Sprintf("Invalid APP_IDLE_TIMEOUT, check it in docker-compose.yml"))
			return errors.New("Invalid APP_IDLE_TIMEOUT")
		}
		GCfg.IdleTimeout = iidleTimeout

		//// Mongo
		// MongoEnable
		mongoEnable := os.Getenv("LOG_MONGO_ENABLE")
		bmongoEnable, err := strconv.ParseBool(mongoEnable)
		if err != nil {
			print(fmt.Sprintf("Invalid LOG_MONGO_ENABLE, check it in docker-compose.yml\n"))
			return errors.New("Invalid LOG_MONGO_ENABLE, non-bool")
		}
		GCfg.MongoEnable = bmongoEnable
		// MongoUrl        string
		mongoUrl := os.Getenv("LOG_MONGO_URL")
		GCfg.MongoUrl = mongoUrl
		//MongoDB          string
		mongoDB := os.Getenv("LOG_MONGO_DATABASE")
		GCfg.MongoDB = mongoDB
		// MongoCollection
		mongoCollection := os.Getenv("LOG_MONGO_COLLECTION")
		GCfg.MongoCollection = mongoCollection
		// MongoWorkerNum
		mongoWorkerNum := os.Getenv("LOG_MONGO_WORKER")
		imongoWorkerNum, err := strconv.Atoi(mongoWorkerNum)
		if err != nil {
			print(fmt.Sprintf("Invalid LOG_MONGO_WORKER, check it in docker-compose.yml\n"))
			return errors.New("Invalid LOG_MONGO_WORKER, non-integer")
		}
		GCfg.MongoWorkerNum = imongoWorkerNum

		//// Redis
		// redisMainflag
		redisMainFlag := os.Getenv("AUTH_REDIS_MAIN")
		iRedisMainFlag, err := strconv.Atoi(redisMainFlag)
		if err != nil {
			print(fmt.Sprintf("Invalid AUTH_REDIS_MAIN, check it in docker-compose.yml\n"))
			return errors.New("Invalid AUTH_REDIS_MAIN")
		}
		GCfg.RedisMainFlag = iRedisMainFlag
		//RedisHost    string
		redisHost := os.Getenv("AUTH_REDIS_MAIN_HOST")
		GCfg.RedisHost = redisHost
		//RedisPort    string
		redisPort := os.Getenv("AUTH_REDIS_MAIN_PORT")
		GCfg.RedisPort = redisPort
		//RedisPwd     string
		redisPwd := os.Getenv("AUTH_REDIS_MAIN_PASSWORD")
		GCfg.RedisPwd = redisPwd
		redisBackupHost := os.Getenv("AUTH_REDIS_BACKUP_HOST")
		GCfg.RedisBackupHost = redisBackupHost
		redisBackupPort := os.Getenv("AUTH_REDIS_BACKUP_PORT")
		GCfg.RedisBackupPort = redisBackupPort
		redisBackupPwd := os.Getenv("AUTH_REDIS_BACKUP_PASSWORD")
		GCfg.RedisBackupPwd = redisBackupPwd
		//RedisReadTimeout
		redisReadTimeout := os.Getenv("AUTH_REDIS_READTIMEOUT")
		iRedisReadTimeout, err := strconv.Atoi(redisReadTimeout)
		if err != nil {
			print(fmt.Sprintf("Invalid AUTH_REDIS_READTIMEOUT in docker-compose.yml\n"))
			return errors.New("Invalid AUTH_REDIS_READTIMEOUT in docker-compose.yml, non-int\n")
		}
		GCfg.RedisReadTimeout = iRedisReadTimeout

		//// Proxy
		//ProxyKeepalive bool
		//ProxyOptions   string
		//ProxyRules     string
		proxyrules := os.Getenv("PROXY_RULES")
		GCfg.ProxyRules = proxyrules
		fmt.Printf("Rules: [%v]\n", proxyrules)

		//// TransformRequest
		//TReqEnable  bool
		//TReqDefault string

		//// TransformResponse
		//TResEnable  bool
		//TResDefault string

		//// Trace
		//TraceEnable  bool
		//TraceDefault string

		//// Quota
		//QuotaEnable bool

		//// Auth
		//AuthEnable      bool
		authEnable := os.Getenv("AUTH_ENABLE")
		bAuthEnable, err := strconv.ParseBool(authEnable)
		if err != nil {
			print(fmt.Sprintf("Invalid AUTH_ENABLE in docker-compose.yml\n"))
			return errors.New("Invalid AUTH_ENABLE in docker-compose.yml")
		}
		GCfg.AuthEnable = bAuthEnable

		//AuthKeepalive   bool
		//AuthOptions     string
		//AuthDefault     string
		//AuthHttpURL     string
		//AuthTimeout     int
		//AuthWhiteIPList string
		//AuthCUSTIDField
		authCustID := os.Getenv("AUTH_CUSTIDFIELD")
		GCfg.AuthCUSTIDField = authCustID
		//AuthCUSTLabelFiled
		authCustLabel := os.Getenv("AUTH_CUSTLABELFIELD")
		GCfg.AuthCUSTLabelField = authCustLabel
		//AuthCUSTIDPingan
		authCUSTIDPingan := os.Getenv("AUTH_CUSTID_PINGAN")
		GCfg.AuthCUSTIDPingan = authCUSTIDPingan

		//// PushMessage
		//PMEnable       bool
		//PMLogpath      string
		//PMPrefix       string
		//PMRedisHost    string
		//PMRedisPort    string
		//PMRedisPwd     string
		//PMRedisChannel string

		//// API
		//ApiEnable bool
		//ApiPort   string
		// 解析黑名单配置
		blacklistConfigName := os.Getenv("BLACKLIST_CONFIG_NAME")
		GCfg.BlacklistConfigName = blacklistConfigName
		iBlacklistConfigTokenMax := os.Getenv("BLACKLIST_CONFIG_CACHE_TOKEN_MAX")
		blacklistConfigTokenMax, err := strconv.Atoi(iBlacklistConfigTokenMax)
		if err != nil {
			fmt.Println("Invalid BLACKLIST_CONFIG_CACHE_TOKEN_MAX, check it in docker-compose.yml")
			return errors.New("解析docker-compose.yml中黑名单配置本地缓存TOKEN最大值错误")
		}
		GCfg.BlacklistConfigTokenMax = blacklistConfigTokenMax
		iBlacklistConfigCustidMax := os.Getenv("BLACKLIST_CONFIG_CACHE_CUSTID_MAX")
		blacklistConfigCustidMax, err := strconv.Atoi(iBlacklistConfigCustidMax)
		if err != nil {
			fmt.Println("Invalid BLACKLIST_CONFIG_CACHE_CUSTID_MAX, check it in docker-compose.yml")
			return errors.New("解析docker-compose.yml中黑名配置单本地缓存CUSTID最大值错误")
		}
		GCfg.BlacklistConfigCustidMax = blacklistConfigCustidMax
		iBlacklistConfigExpiration := os.Getenv("BLACKLIST_CONFIG_TOKEN_EXPIRATION_TIME")
		blacklistConfigExpiration, err := strconv.ParseInt(iBlacklistConfigExpiration, 10, 64)
		if err != nil {
			fmt.Println("Invalid BLACKLIST_CONFIG_TOKEN_EXPIRATION_TIME, check it in docker-compose.yml")
			return errors.New("解析docker-compose.yml中黑名单配置TOEKN本地缓存过期时间错误")
		}
		GCfg.BlacklistConfigTokenExpiration = blacklistConfigExpiration
	}

	return nil
}
