package hardwareinfocollector

import (
	// "bufio"
	"os/exec"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var (
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
			Help: "Total number of memory slots",
		},
	)
)

func init() {
	prometheus.MustRegister(memorySlotGauge, memoryTotalSlotsGauge)
	collectMemoryDMI()
}

func collectMemoryDMI() {
	cmd := exec.Command("dmidecode", "-t", "17")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	// scanner := bufio.NewScanner(strings.NewReader(string(output)))
	// var currentSlot, size, typ, speed, manu, part string
	// totalSlots := 0

	// for scanner.Scan() {
	// 	line := strings.TrimSpace(scanner.Text())
	// 	if strings.HasPrefix(line, "Memory Device") {
	// 		// 新内存条开始
	// 		currentSlot = ""
	// 		size = "0"
	// 		typ = "unknown"
	// 		speed = "0"
	// 		manu = "unknown"
	// 		part = "unknown"
	// 		totalSlots++
	// 	} else if strings.HasPrefix(line, "Locator:") {
	// 		currentSlot = strings.TrimSpace(strings.TrimPrefix(line, "Locator:"))
	// 	} else if strings.HasPrefix(line, "Size:") {
	// 		if strings.Contains(line, "No Module Installed") {
	// 			continue // 跳过空插槽
	// 		}
	// 		sizeStr := strings.TrimSpace(strings.TrimPrefix(line, "Size:"))
	// 		if strings.HasSuffix(sizeStr, "GB") {
	// 			val, _ := strconv.ParseFloat(strings.TrimSuffix(sizeStr, " GB"), 64)
	// 			size = strconv.FormatInt(int64(val*1e9), 10)
	// 		} else if strings.HasSuffix(sizeStr, "MB") {
	// 			val, _ := strconv.ParseFloat(strings.TrimSuffix(sizeStr, " MB"), 64)
	// 			size = strconv.FormatInt(int64(val*1e6), 10)
	// 		}
	// 	} else if strings.HasPrefix(line, "Type:") {
	// 		typ = strings.TrimSpace(strings.TrimPrefix(line, "Type:"))
	// 	} else if strings.HasPrefix(line, "Speed:") {
	// 		speedStr := strings.TrimSpace(strings.TrimPrefix(line, "Speed:"))
	// 		if strings.HasSuffix(speedStr, "MT/s") {
	// 			speed = strings.TrimSuffix(speedStr, " MT/s")
	// 		}
	// 	} else if strings.HasPrefix(line, "Manufacturer:") {
	// 		manu = strings.TrimSpace(strings.TrimPrefix(line, "Manufacturer:"))
	// 	} else if strings.HasPrefix(line, "Part Number:") {
	// 		part = strings.TrimSpace(strings.TrimPrefix(line, "Part Number:"))
	// 	}
	// }

	// 实际上上面逻辑不完整（需按块解析），这里简化：只处理已填充的插槽
	// 更健壮的做法是按空行分段，但为简洁起见，假设每遇到 "Memory Device" 就重置

	// 重新解析（更可靠方式）
	blocks := strings.Split(string(output), "\n\n")
	populated := 0
	for _, block := range blocks {
		if !strings.Contains(block, "Memory Device") {
			continue
		}
		lines := strings.Split(block, "\n")
		slot := "unknown"
		sizeBytes := "0"
		memType := "unknown"
		speedMhz := "0"
		manufacturer := "unknown"
		partNumber := "unknown"
		populatedFlag := false

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Locator:") {
				slot = strings.TrimSpace(strings.TrimPrefix(line, "Locator:"))
			} else if strings.HasPrefix(line, "Size:") {
				sizeStr := strings.TrimSpace(strings.TrimPrefix(line, "Size:"))
				if !strings.Contains(sizeStr, "No Module Installed") {
					populatedFlag = true
					if strings.HasSuffix(sizeStr, "GB") {
						val, _ := strconv.ParseFloat(strings.TrimSuffix(sizeStr, " GB"), 64)
						sizeBytes = strconv.FormatInt(int64(val*1e9), 10)
					} else if strings.HasSuffix(sizeStr, "MB") {
						val, _ := strconv.ParseFloat(strings.TrimSuffix(sizeStr, " MB"), 64)
						sizeBytes = strconv.FormatInt(int64(val*1e6), 10)
					}
				}
			} else if strings.HasPrefix(line, "Type:") {
				memType = strings.TrimSpace(strings.TrimPrefix(line, "Type:"))
			} else if strings.HasPrefix(line, "Speed:") {
				speedStr := strings.TrimSpace(strings.TrimPrefix(line, "Speed:"))
				if speedStr != "Unknown" {
					speedMhz = strings.Fields(speedStr)[0] // 取数字部分
				}
			} else if strings.HasPrefix(line, "Manufacturer:") {
				manufacturer = strings.TrimSpace(strings.TrimPrefix(line, "Manufacturer:"))
			} else if strings.HasPrefix(line, "Part Number:") {
				partNumber = strings.TrimSpace(strings.TrimPrefix(line, "Part Number:"))
			}
		}

		if populatedFlag {
			populated++
			memorySlotGauge.WithLabelValues(
				slot, sizeBytes, memType, speedMhz, manufacturer, partNumber,
			).Set(1)
		}
	}

	memoryTotalSlotsGauge.Set(float64(populated)) // 或 totalSlots（包括空的）
}