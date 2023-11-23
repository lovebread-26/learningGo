package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"api-gw/blacklist"
	"api-gw/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
)

// Data-struct >>>>>
/*
{
    "expireAt": 1690167766,
    "data": {
        "id": "cfbf261b68ed472ca623524fffe765d32",
        "data": {
            "_id": "5ffd4b8699acd4eef0e9babe",
            "account": "aly21",
            "bucket": "ctsi",
            "bucket_id": 2,
            "nickname": "aly21",
            "mobile": "19112345678",
            "email": "11@qq.com",
            "remark": "11@qq.com",
            "uid": "542bdd7054a511ebbad3053dde46adb5",
            "create_time": "2021-01-12 15:11:02",
            "status": "1",
            "source": "0",
            "rid": 10,
            "authLockExp": 0,
            "lastLoginIp": "127.0.0.1",
            "lastLoginTime": 1689835941786,
            "pwdErrNum": 0,
            "cust_id": "00002121",
            "customer_id": "00002121",
            "custId": "custIdcustId",
            "sipuri": "sip:19112345678@192.168.129.2",
            "skip_sms_capt": true,
            "completeStatus_bak": {
                "aa": "13213"
            },
            "entprise_city": "\xe5\x90\x88\xe8\x82\xa5\xe5\xb8\x82",
            "entprise_province": "\xe5\xae\x89\xe5\xbe\xbd\xe7\x9c\x81",
            "auth": [

            ],
            "quota": [
                "rule_test"
            ],
            "title": "bininsert",
            "trace": [
                "mongodb",
                "http"
            ],
            "transformRequest": [

            ],
            "url": "http://127.0.0.1:3533/etd/api/dev189/bininsert",
            "apply": {
                "app": "chatbot",
                "app_id": "cfbf261b68ed472ca623524fffe765d32",
                "app_key": "2cd84f9d44524520f6cd1cf220c520921",
                "app_secret": "4113a0f48be27903e5121b13b628d6f2a",
                "whiteListIP_1": "192.168.11.11"
            }
        },
        "isAdmin": false,
        "allowMultiLogin": false
    }
}
*/
// <<<<<<<
type dataDetail struct {
	Id                string `json:"_id"`
	Account           string `json:"account"`
	Bucket            string `json:"bucket"`
	BucketId          int    `json:"bucket_id"`
	NickName          string `json:"nickname"`
	Mobile            string `json:"mobile"`
	Email             string `json:"email"`
	Remark            string `json:"remark"`
	Uid               string `json:"uid"`
	CreateTime        string `json:"create_time"`
	Status            string `json:"status"`
	Source            string `json:"source"`
	Rid               int    `json:"rid"`
	AuthLockExp       int    `json:"authLockExp"`
	LastLoginIp       string `json:"lastLoginIp"`
	LastLoginTime     int    `json:"lastLoginTime"`
	PwdErrNum         int    `json:"pwdErrNum"`
	CustId            string `json:"cust_id"`
	CustomerId        string `json:"customer_id"`
	CustIdNoUnderline string `json:"custId"`
	CustLabel         string `json:"label"`
	SipUri            string `json:"sipuri"`
	SkipSMSCapt       bool   `json:"skip_sms_capt"`
	//CompleteStatusBak struct
	EntpriseCity    string `json:"entprise_city"`
	EntpriseProvice string `json:"entprise_province"`
	//Auth struct
	//Quota struct
	Title string `json:"title"`
	//Trace struct
	//TransformRequest struct
	Url   string      `json:"url"`
	Apply applyDetail `json:"apply"`
}

type applyDetail struct {
	App         string `json:"app"`
	AppId       string `json:"app_id"`
	AppKey      string `json:"app_key"`
	AppSecret   string `json:"app_secret"`
	WhiteListIP string `json:"whiteListIP"`
}

type inerData struct {
	Id   string     `json:"id"`
	Data dataDetail `json:"data"`
}

type authData struct {
	ExpireAt        int64    `json:"expireAt"`
	Data            inerData `json:"data"`
	IsAdmin         bool     `json:"isAdmin"`
	AllowMultiLogin bool     `json:"allowMultiLogin"`
}

// Get the accessToken from
// 1) query parameter; ?access_token=xxxxxxx
// 2) Authorization: Bearer xxxxxxxx
func getAccessTokenByRequest(c *gin.Context) string {
	access_token := ""
	// TODO  Get the access_token from Authorization
	//hAuth := c.Request.Header.Get("Authorization")
	// authorization.indexOf('Bearer') === 0 then regex /\S+$/[0]
	access_token = c.Query("access_token")
	//access_token = c.Request.Header.Get("access_token")
	return access_token
}

// Get data from redis by access_token
func getData(redisHandle *redis.Client, at string) (data string) {
	if at == "" {
		return
	}
	// Prefix
	// Prefix: dev-op:AccessToken:xxxxxxx (xxxxxxx: token)
	akey := "dev-op:AccessToken:" + at
	val, err := redisHandle.Get(akey).Result()
	if err != nil {
		fmt.Printf("Failed to get data from redis by access_token [%v] with error [%v]\n", akey, err.Error())
		return
	}
	data = val

	// Redis handler
	return
}

// auth
func AuthHandle() gin.HandlerFunc {
	return func(c *gin.Context) {
		nowTime := time.Now()

		//gMetrics, gok := c.Get("gMetrics")
		// Update the traceLog
		traceLog, ok := c.Get("traceLogObj")
		gwm, gok := c.Get("gwm")
		if !gok {
			utils.LoggerIt(c, "error", "gwm(GwMetrics) is none!!!!!!")
		}
		gwcm, gwok := c.Get("gwcm")
		if !gwok {
			utils.LoggerIt(c, "error", "gwcm is none!!!!!!")
			// TODO logger
		}
		var traceLogObj *utils.TraceLog
		if !ok {
			utils.LoggerIt(c, "error", "traceLogObj is none!!!!!!!!!!!!!!")
		} else {
			traceLogObj = traceLog.(*utils.TraceLog)
		}

		// 1) Get the access_token
		access_token := getAccessTokenByRequest(c)
		if access_token == "" {
			c.Abort()
			errMsg := "Found no access_token"
			if traceLogObj != nil {
				traceLogObj.ErrorMsg = errMsg
				// Logged in preHandle
				//utils.LoggerIt(c, "error", "")
			} else {
				utils.LoggerIt(c, "error", errMsg)
			}
			if gok {
				//atomic.AddInt64(&gMetrics.(*utils.BaseData).SendFailNum, 1)
				//TODO metrics fail
				if gwm != nil {
					gwm.(*prometheus.CounterVec).WithLabelValues("sendFail").Inc()
				}
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"code": "20003",
				"msg":  errMsg,
			})
			return
		}

		// 1.1）查询token是否在本地黑名单缓存中
		// 如果在黑名单中，禁止继续转发
		// 如果不在黑名单中，继续后面的流程
		if ok := blacklist.SearchBlacklistByToken(access_token); ok {
			c.Abort()
			fmt.Println("token", access_token, "在黑名单中，禁止转发")
			errMsg := fmt.Sprintf("access_token %s 在黑名单中，禁止转发", access_token)
			utils.LoggerIt(c, "warn", errMsg)
			c.String(http.StatusUnauthorized, errMsg)
			return
		}

		// 2) Get the value from redis
		// Get the redisHandle
		var tryFlag = false
		var handle *redis.Client
		var redisMainFlag = 1
		redisMainFlagPointer, ok := c.Get("redisMainFlag")
		if !ok {
			redisMainFlag = 1
		} else {
			redisMainFlag = *redisMainFlagPointer.(*int)
		}

		redisHandle, ok := c.Get("redisHandle")
		if !ok || redisHandle == "" || redisHandle.(*redis.Client) == nil {
			if redisMainFlag == 1 {
				redisMainFlag = 0
			}
			tryFlag = true
			//c.Abort()
			//errMsg := "Redis未连接"
			//if traceLogObj != nil {
			//	traceLogObj.ErrorMsg = errMsg
			//	// Logged in preHandle
			//	//utils.LoggerIt(c, "error", "")
			//} else {
			//	utils.LoggerIt(c, "error", errMsg)
			//}
			//if gok {
			//	atomic.AddInt64(&gMetrics.(*utils.BaseData).SendFailNum, 1)
			//	if gwm != nil {
			//		gwm.(*prometheus.CounterVec).WithLabelValues("sendFail").Inc()
			//	}
			//}
			//c.JSON(http.StatusBadRequest, gin.H{
			//	"code": "20002",
			//	"msg":  errMsg,
			//})
			//return
		}

		redisBackupHandle, ok := c.Get("redisBackupHandle")
		if !ok || redisBackupHandle == "" || redisBackupHandle.(*redis.Client) == nil {
			if redisMainFlag == 0 {
				if !tryFlag {
					redisMainFlag = 1
					tryFlag = true
				} else {
					c.Abort()
					errMsg := "Redis未连接"
					if traceLogObj != nil {
						traceLogObj.ErrorMsg = errMsg
						// Logged in preHandle
						//utils.LoggerIt(c, "error", "")
					} else {
						utils.LoggerIt(c, "error", errMsg)
					}
					if gok {
						//atomic.AddInt64(&gMetrics.(*utils.BaseData).SendFailNum, 1)
						if gwm != nil {
							gwm.(*prometheus.CounterVec).WithLabelValues("sendFail").Inc()
						}
					}
					c.JSON(http.StatusBadRequest, gin.H{
						"code": "20002",
						"msg":  errMsg,
					})
					return
				}
			}
		}

		if redisMainFlag == 0 {
			handle = redisBackupHandle.(*redis.Client)
		} else {
			handle = redisHandle.(*redis.Client)
		}

		//data := getData(redisHandle.(*redis.Client), access_token)
	TRYIT:
		data := getData(handle, access_token)
		if data == "" {
			if !tryFlag {
				if redisMainFlag == 1 {
					// switch to backup
					handle = redisBackupHandle.(*redis.Client)
					// Update the mainFlag
					*redisMainFlagPointer.(*int) = 0
				} else {
					// switch to main
					handle = redisHandle.(*redis.Client)
					// Update the mainFlag
					*redisMainFlagPointer.(*int) = 1
				}
				tryFlag = true
				goto TRYIT
			} else {
				c.Abort()
				errMsg := "没有找到和access_token匹配的数据"
				if traceLogObj != nil {
					traceLogObj.ErrorMsg = errMsg
					// Logged in preHandle
					//utils.LoggerIt(c, "error", "")
				} else {
					utils.LoggerIt(c, "error", errMsg)
				}
				if gok {
					//atomic.AddInt64(&gMetrics.(*utils.BaseData).SendFailNum, 1)
					if gwm != nil {
						gwm.(*prometheus.CounterVec).WithLabelValues("sendFail").Inc()
					}
				}
				c.String(http.StatusUnauthorized, errMsg)
				return
			}
		}

		if data == "" && tryFlag {
			c.Abort()
			errMsg := "没有找到和access_token匹配的数据"
			if traceLogObj != nil {
				traceLogObj.ErrorMsg = errMsg
				// Logged in preHandle
				//utils.LoggerIt(c, "error", "")
			} else {
				utils.LoggerIt(c, "error", errMsg)
			}
			if gok {
				//atomic.AddInt64(&gMetrics.(*utils.BaseData).SendFailNum, 1)
				if gwm != nil {
					gwm.(*prometheus.CounterVec).WithLabelValues("sendFail").Inc()
				}
			}
			c.String(http.StatusUnauthorized, errMsg)
			return
		}

		// Update AuthInfo
		if traceLogObj != nil {
			traceLogObj.AuthInfo = data
		}

		// Parse the data
		aData := new(authData)
		err := json.Unmarshal([]byte(data), &aData)
		if err != nil {
			c.Abort()
			if traceLogObj != nil {
				traceLogObj.ErrorMsg = err.Error()
				// Logged in preHandle
				//utils.LoggerIt(c, "error", "")
			} else {
				utils.LoggerIt(c, "error", err.Error())
			}
			if gok {
				//atomic.AddInt64(&gMetrics.(*utils.BaseData).SendFailNum, 1)
				if gwm != nil {
					gwm.(*prometheus.CounterVec).WithLabelValues("sendFail").Inc()
				}
			}
			c.JSON(http.StatusBadRequest, gin.H{
				"code": "20003",
				"msg":  err.Error(),
			})
			return
		}

		// ClientId, ClientLabel
		custId := ""
		switch utils.GCfg.AuthCUSTIDField {
		case "cust_id":
			if aData.Data.Data.CustId != "" {
				custId = aData.Data.Data.CustId
			}
		case "custId":
			if aData.Data.Data.CustIdNoUnderline != "" {
				custId = aData.Data.Data.CustIdNoUnderline
			}
		case "customer_id":
			if aData.Data.Data.CustomerId != "" {
				custId = aData.Data.Data.CustomerId
			}
		}
		if custId != "" {
			c.Set("CustId", custId)
		}

		// 3）查询custId是否在本地黑名单缓存中
		// 如果在黑名单中，禁止继续转发
		// 如果不在黑名单中，继续后面的流程
		if ok := blacklist.SearchBlacklistByCustId(custId, access_token); ok {
			c.Abort()
			fmt.Println("custId", custId, "在黑名单中，禁止转发")
			errMsg := fmt.Sprintf("custId %s 在黑名单中，禁止转发", custId)
			utils.LoggerIt(c, "warn", errMsg)
			c.String(http.StatusUnauthorized, errMsg)
			return
		}

		// TODO ClientLabel
		if aData.Data.Data.CustLabel != "" {
			custId = aData.Data.Data.CustLabel
			c.Set("CustLabel", aData.Data.Data.CustLabel)
		}

		// App_id AppId
		//appId := aData.Data.Data.Apply.AppId
		//if appId != "" {
		//	// Record AppId
		//	c.Set("AppId", appId)
		//}

		// TODO WhiteListIP
		if aData.Data.Data.Apply.WhiteListIP != "" {
			if !strings.Contains(aData.Data.Data.Apply.WhiteListIP, c.ClientIP()) {
				c.Abort()
				if gok {
					//atomic.AddInt64(&gMetrics.(*utils.BaseData).SendFailNum, 1)
					if gwm != nil {
						gwm.(*prometheus.CounterVec).WithLabelValues("sendFail").Inc()
					}
				}
				if gwok {
					if gwcm != nil {
						//gwcm.(*prometheus.CounterVec).WithLabelValues("clientSendFail", appId).Inc()
						gwcm.(*prometheus.CounterVec).WithLabelValues("clientSendFail", custId).Inc()
					}
				}
				errMsg := c.ClientIP() + " not match the WHITE-IP-LIST [" + aData.Data.Data.Apply.WhiteListIP + "]"
				if traceLogObj != nil {
					traceLogObj.ErrorMsg = errMsg
					// Logged in preHandle
					//utils.LoggerIt(c, "error", "")
				} else {
					utils.LoggerIt(c, "error", errMsg)
				}
				c.String(http.StatusUnauthorized, "权限错误:无权访问(1003)")
				return
			}
		}

		c.Next()
		// TODO
		costTime := time.Since(nowTime)
		url := c.Request.URL.String()
		print(fmt.Sprintf("AuthHandle cost %v for url [%v]\n", costTime, url))
	}
}
