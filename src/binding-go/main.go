package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

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

func Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// time.Sleep(10 * time.Second)

		handleRequestBody(c)

		fmt.Printf("Request : [%v]\n", c.Request)
		fmt.Printf("Request-body : [%v]\n", c.Request.Body)

		//selectData := `{"returnCode":"0","msg":"成功","resBody":{"toMongoList":[{"numberA":"18811595237","numberX":"18811595237","serviceId":"18811595237"},{"numberA":"19902290012","numberX":"19902290013","serviceId":"djslakdjla"}],"existList":[{"numberA":"18211223000","numberX":"18211223000","serviceId":"djslakdjla"},{"numberA":"18232112765","numberX":"18232112765","serviceId":"djslakdjla"}],"inaccuracyList":[{"numberA":"19902290012","numberX":"19902290013","serviceId":"djslakdjla"},{"numberA":"19902290012","numberX":"19902290013","serviceId":"djslakdjla"}]}}`
		deleteData := `{"msg":"成功！","returnCode":"0","resBody":[{"serviceId":"888859a205a34f45acb84d621dc29999","numberA":"19910980012","numberX":"19910980011"},{"serviceId":"54daaxa205a34f112dhdsa63scc2sacw","numberA":"19910980012","numberX":"19910980013"}]}`
		//jsonData, err := json.Marshal([]byte(data))
		//if err != nil {
		//	fmt.Printf("Failed to marshal to string with error [%s]\n", err.Error())
		//}

		//c.Data(http.StatusOK, "application/json", []byte(selectData))
		c.Data(http.StatusOK, "application/json", []byte(deleteData))
	}
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	ginlogf, err := os.Create("gin.log")
	if err != nil {
		fmt.Printf("Failed to create gin.log with error [%s]\n", err.Error())
	}
	gin.DefaultErrorWriter = io.MultiWriter(ginlogf)

	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	router.Any("/*proxyPath", Handler())

	if err = router.Run(":8000"); err != nil {
		fmt.Printf("Failed to run bind-server with port 8000 with error [%s]\n", err.Error())
	}

	fmt.Printf("Bind-Server is running. Listen on port :8000\n")
}
