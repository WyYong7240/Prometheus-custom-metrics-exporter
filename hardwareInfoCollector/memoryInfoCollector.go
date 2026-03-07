package hardwareinfocollector

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// 使用 Describe/Collect 模式或在采集时 Reset 以防止旧数据残留
	memorySlotGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "hardware_memory_slot_info",
			Help: "Memory slot information (value=1 if populated)",
		},
		[]string{"slot", "size_bytes", "type", "speed_mhz", "manufacturer", "part_number"},
	)
	memoryTotalSlotsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "hardware_memory_slots_total",
			Help: "Total number of memory slots detected",
		},
	)
)

func init() {
	prometheus.MustRegister(memorySlotGauge, memoryTotalSlotsGauge)
}

// UpdateMemoryMetrics 建议在 Prometheus 采集请求到达时调用此函数
func UpdateMemoryMetrics() {
	// 每次采集前清空旧指标，防止插槽物理变动后旧数据依然存在
	memorySlotGauge.Reset()

	output, err := exec.Command("dmidecode", "-t", "17").Output()
	if err != nil {
		return
	}

	// 按照 DMI 结构体块分割
	blocks := strings.Split(string(output), "Memory Device")
	totalSlots := 0
	populatedSlots := 0

	for _, block := range blocks {
		// 跳过非设备块（如 header）
		if !strings.Contains(block, "Size:") {
			continue
		}

		totalSlots++
		data := parseDmiBlock(block)

		// 检查是否有安装模块
		sizeStr := data["Size"]
		if sizeStr == "" || strings.Contains(sizeStr, "No Module Installed") {
			continue
		}

		populatedSlots++

		// 转换单位并上报
		sizeBytes := parseSizeToBytes(sizeStr)
		speed := strings.Fields(data["Speed"])[0] // 提取 "3200 MT/s" 中的 "3200"

		memorySlotGauge.WithLabelValues(
			data["Locator"],
			strconv.FormatInt(sizeBytes, 10),
			data["Type"],
			speed,
			data["Manufacturer"],
			data["Part Number"],
		).Set(1)
	}

	memoryTotalSlotsGauge.Set(float64(totalSlots))
}

// 辅助函数：将 DMI 块解析为 Map
func parseDmiBlock(block string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(block, "\n")
	for _, line := range lines {
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		result[key] = val
	}
	return result
}

// 辅助函数：处理内存单位转换为 Bytes (使用 1024 进制更符合 OS 习惯)
func parseSizeToBytes(sizeStr string) int64 {
	re := regexp.MustCompile(`(\d+)\s*(GB|MB|KB|B)`)
	match := re.FindStringSubmatch(sizeStr)
	if len(match) < 3 {
		return 0
	}

	val, _ := strconv.ParseInt(match[1], 10, 64)
	unit := match[2]

	switch unit {
	case "GB":
		return val * 1024 * 1024 * 1024
	case "MB":
		return val * 1024 * 1024
	case "KB":
		return val * 1024
	default:
		return val
	}
}
