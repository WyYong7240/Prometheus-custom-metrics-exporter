package systemcollector

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// SystemCollector 定义系统信息的采集器
type SystemCollector struct {
	userInfo    *prometheus.Desc
	appInfo     *prometheus.Desc
	hostEtcPath string // 宿主机的 /etc 挂载路径，比如 "/host/etc"
}

// NewSystemCollector 实例化 Collector
func NewSystemCollector(hostEtcPath string) *SystemCollector {
	return &SystemCollector{
		// 暴露用户和用户组信息，值为固定 1，信息存在 label 里
		userInfo: prometheus.NewDesc(
			"system_user_group_info",
			"Mapping of users and their groups on the host",
			[]string{"username", "groupname"},
			nil,
		),
		// 暴露已安装的应用信息
		appInfo: prometheus.NewDesc(
			"system_installed_app_info",
			"Installed applications and their versions on the host",
			[]string{"app_name", "version"},
			nil,
		),
		hostEtcPath: hostEtcPath,
	}
}

// Describe 实现 prometheus.Collector 接口
func (c *SystemCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.userInfo
	ch <- c.appInfo
}

// Collect 实现 prometheus.Collector 接口
func (c *SystemCollector) Collect(ch chan<- prometheus.Metric) {
	c.collectUsersAndGroups(ch)
	c.collectInstalledApps(ch)
}

// collectUsersAndGroups 解析 /etc/passwd 和 /etc/group
func (c *SystemCollector) collectUsersAndGroups(ch chan<- prometheus.Metric) {
	// 注意：实际生产中需要解析 /host/etc/passwd 和 /host/etc/group 进行关联映射
	// 这里为了演示，采用简化版的逻辑，直接读取 passwd 中的用户
	passwdFile := c.hostEtcPath + "/passwd"
	file, err := os.Open(passwdFile)
	if err != nil {
		// 读取失败时记录日志或忽略，不要让整个 collector 崩溃
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) >= 4 {
			username := parts[0]
			gid := parts[3]
			// 在真实的完整实现中，你需要读取 /etc/group 文件把 gid 翻译成真实的 groupname
			// 这里假设我们已经翻译好了
			groupname := "group_" + gid

			// 输出指标
			ch <- prometheus.MustNewConstMetric(
				c.userInfo,
				prometheus.GaugeValue,
				1.0, // 固定值为 1
				username, groupname,
			)
		}
	}
}

// collectInstalledApps 获取已安装的软件列表
func (c *SystemCollector) collectInstalledApps(ch chan<- prometheus.Metric) {
	// 注意 1：需要针对宿主机的包管理器（apt/dpkg 或 yum/rpm）执行命令。
	// 注意 2：容器内部需要有 chroot 权限，或者通过 nsenter 借用宿主机的命名空间执行命令。

	// 假设宿主机是 Debian/Ubuntu，我们在容器内通过 chroot 到宿主机根目录执行 dpkg-query
	cmd := exec.Command("chroot", "/host", "dpkg-query", "-W", "-f=${Package}::${Version}\n")
	out, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "::")
		if len(parts) == 2 {
			appName := parts[0]
			version := parts[1]

			ch <- prometheus.MustNewConstMetric(
				c.appInfo,
				prometheus.GaugeValue,
				1.0,
				appName, version,
			)
		}
	}
}
