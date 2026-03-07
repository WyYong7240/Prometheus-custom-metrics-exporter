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
	// 1. 初始化收集各个指标信息 (注册到 Prometheus)
	// 内存指标在 hardwareinfocollector 的 init() 函数中已经注册过了
	// 我们先进行一次初始采集
	hardwareinfocollector.CollectCPUInfo()
	hardwareinfocollector.UpdateMemoryMetrics()

	// 2. 设置 HTTP 路由
	// Prometheus 抓取端点
	http.Handle("/metrics", promhttp.Handler())

	// 设置手动触发更新的路由
	http.HandleFunc("/refresh", func(w http.ResponseWriter, r *http.Request) {
		hardwareinfocollector.CollectCPUInfo()
		hardwareinfocollector.UpdateMemoryMetrics() // 调用新修复的内存采集函数
		w.Write([]byte("Hardware info (CPU & Memory) refreshed"))
	})

	// 业务逻辑路由
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()
		httprequesttotal.RequestCounter.Inc()
		httprequesttotal.RequestDuration.Observe(time.Since(start).Seconds())
		writer.Write([]byte("Hello, Prometheus!"))
	})

	// 3. (可选) 定时自动刷新硬件信息，防止长期运行数据过时
	go func() {
		for {
			time.Sleep(5 * time.Minute) // 每5分钟更新一次硬件状态
			hardwareinfocollector.UpdateMemoryMetrics()
		}
	}()

	fmt.Println("Starting server on :30012")
	fmt.Println("Access /metrics for Prometheus metrics")
	fmt.Println("Access /refresh to manually refresh hardware info")

	err := http.ListenAndServe(":30012", nil)
	if err != nil {
		fmt.Printf("Server failed: %s\n", err)
	}
}
