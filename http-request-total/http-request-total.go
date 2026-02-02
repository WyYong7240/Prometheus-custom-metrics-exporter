package httprequesttotal

import (
	"github.com/prometheus/client_golang/prometheus"
)

// 每当有人访问/路径时，记录一次请求，并统计请求耗时
var (
	// 请求计数器
	RequestCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "myapp_requests_total",
		Help: "Total number of requests received.",
	})
	// 请求延迟直方图,统计每次请求的耗时
	RequestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "myapp_request_duration_seconds",
		Help: "Duration of requests in seconds",
		// 线性分布的桶，从0.1秒开始，每个桶间隔0.1秒，共10个桶
		Buckets: prometheus.LinearBuckets(0.1, 0.1, 10),
	})
)

// 将自定义的指标注册到Prometheus的默认注册表中
func init() {
	prometheus.MustRegister(RequestCounter)
	prometheus.MustRegister(RequestDuration)
}
