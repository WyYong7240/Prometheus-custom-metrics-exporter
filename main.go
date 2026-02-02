package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	hardwareinfocollector "github.com/wuyong7240/prometheus-custom-metrics/hardwareInfoCollector"
	httprequesttotal "github.com/wuyong7240/prometheus-custom-metrics/http-request-total"
)

func main() {
	// WYB
	// 初始化收集各个指标信息
	hardwareinfocollector.CollectCPUInfo()

	// 设置 HTTP 路由
	http.Handle("/metrics", promhttp.Handler())

	// 设置触发所有信息更新的路由
	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		hardwareinfocollector.CollectCPUInfo()
		w.Write([]byte("CPU info refreshed"))
	})

	// 当访问根路径/时,设置访问数量增加
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()                                                   // 记录当前时间
		httprequesttotal.RequestCounter.Inc()                                 // 调用INc方法，增加请求计数
		httprequesttotal.RequestDuration.Observe(time.Since(start).Seconds()) // 调用Observe记录请求耗时
		writer.Write([]byte("Hello, Prometheus!"))                            // 返回响应内容
	})

	fmt.Println("Starting server on :30100")
	fmt.Println("Access /metrics for Prometheus metrics")
	fmt.Println("Access /refresh to refresh CPU info")
	http.ListenAndServe(":30100", nil)
}
