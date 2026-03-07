# --- 第一阶段：构建 (Build Stage) ---
# 修改点：改用腾讯云镜像源 (CCR)，解决阿里云源权限问题
FROM ccr.ccs.tencentyun.com/library/golang:1.21-alpine AS builder

# 1. 设置 Go 代理以加速国内下载
ENV GOPROXY=https://goproxy.cn,direct

# 2. 设置工作目录
WORKDIR /app

# 3. 首先复制依赖文件（利用 Docker 缓存层）
COPY go.mod go.sum ./
RUN go mod download

# 4. 复制整个项目源码
# 确保 hardwareInfoCollector, http-request-total, main.go 都被复制进去
COPY . .

# 5. 编译
# -ldflags "-s -w" 可以减小二进制体积
# CGO_ENABLED=0 确保生成的二进制文件在 Alpine 中不需要额外的 C 库
RUN CGO_ENABLED=0 GOOS=linux go build -o hardware-exporter ./main.go

# --- 第二阶段：运行 (Run Stage) ---
# 修改点：改用腾讯云镜像源 (CCR)
FROM ccr.ccs.tencentyun.com/library/alpine:latest

# 1. 替换 Alpine apk 源为阿里云源，并安装所需工具
# dmidecode 是核心工具，ca-certificates 确保 HTTPS 请求正常
# 这一步非常关键，防止 apk add 时连接官方源超时
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache dmidecode ca-certificates

# 2. 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/hardware-exporter /usr/local/bin/hardware-exporter

# 3. 暴露应用端口
EXPOSE 30012

# 4. 启动程序
# 提示：在 K8s 中部署时必须开启 privileged: true 才能让 dmidecode 正常工作
ENTRYPOINT ["/usr/local/bin/hardware-exporter"]