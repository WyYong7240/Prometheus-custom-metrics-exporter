package hardwareinfocollector

import (
	"fmt"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v3/cpu"
)

// 自定义指标
var (
	cpuModelGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hardware_cpu_model_info",
			Help: "CPU model information, value is always 1, labels contain details",
		},
		[]string{"model", "vendor", "architecture"},
	)

	cpuCoresGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "hardware_cpu_cores_total",
			Help: "Total number of CPU cores",
		},
	)

	cpuThreadsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "hardware_cpu_threads_total",
			Help: "Total number of CPU threads",
		},
	)

	cpuFrequencyGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "hardware_cpu_frequency_hertz",
			Help: "CPU frequency in Hz",
		},
	)
)

func init() {
	// 注册指标
	prometheus.MustRegister(cpuModelGauge)
	prometheus.MustRegister(cpuCoresGauge)
	prometheus.MustRegister(cpuThreadsGauge)
	prometheus.MustRegister(cpuFrequencyGauge)
}

// 收集 CPU 信息
func CollectCPUInfo() {
	// 获取 CPU 信息统计
	info, err := cpu.Info()
	if err != nil {
		fmt.Printf("Error getting CPU info: %v\n", err)
		return
	}

	if len(info) > 0 {
		cpuInfo := info[0] // 第一个 CPU 的信息

		// 设置 CPU 型号信息
		var modelName, vendor string
		var mhz float64
		modelName = cpuInfo.ModelName
		vendor = cpuInfo.VendorID
		mhz = cpuInfo.Mhz

		// 获取物理核心数
		physicalCores, err := cpu.Counts(false)
		if err != nil {
			fmt.Printf("Error getting physical core count: %v\n", err)
			physicalCores = 0
		}

		// 获取逻辑核心数，即线程数
		logicalCores, err := cpu.Counts(true)
		if err != nil {
			fmt.Printf("Error getting logical core count: %v\n", err)
			logicalCores = 0
		}

		// 设置指标
		cpuModelGauge.WithLabelValues(modelName, vendor, runtime.GOARCH).Set(1)
		cpuCoresGauge.Set(float64(physicalCores))
		cpuFrequencyGauge.Set(mhz * 1e6)	// 转换为Hz
		cpuThreadsGauge.Set(float64(logicalCores))
	}
}
