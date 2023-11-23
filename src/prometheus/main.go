package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 注册Prometheus指标
	reg := prometheus.NewRegistry()

	// 创建自定义指标
	counter := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "my_counter",
		Help: "This is my counter.",
	}, []string{"label"})

	// 注册指标到Registry
	reg.MustRegister(counter)

	// 启动HTTP服务以提供Prometheus指标数据
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}).ServeHTTP(w, r)
	})

	// 模拟一些数据更新操作
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		counter.WithLabelValues("value1").Inc()
		counter.WithLabelValues("value2").Inc()
		counter.WithLabelValues("value3").Inc()
	}

	http.ListenAndServe(":8080", nil)
}
