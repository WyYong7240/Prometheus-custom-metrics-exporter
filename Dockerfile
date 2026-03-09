# --- 第一阶段：构建 (Build Stage) ---
# 使用南京大学 (NJU) 镜像源，这是目前国内最稳定的加速源之一
FROM docker.m.daocloud.io/library/golang:1.24-alpine AS builder

# 1. 设置 Go 代理以加速国内依赖下载
ENV GOPROXY=https://goproxy.cn,direct

# 2. 设置工作目录
WORKDIR /app

# 3. 首先复制依赖文件（利用 Docker 缓存层，只要 mod/sum 没变就不会重新下载依赖）
COPY go.mod go.sum ./
RUN go mod download

# 4. 复制整个项目源码
COPY . .

# 5. 编译
# -ldflags "-s -w" 移除符号表和调试信息，显著减小二进制体积
# CGO_ENABLED=0 确保静态链接，使其能在精简的 Alpine 镜像中运行
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o hardware-exporter ./main.go

# --- 第二阶段：运行 (Run Stage) ---
FROM docker.m.daocloud.io/library/alpine:latest

# 1. 替换 Alpine apk 源为阿里云源，并安装所需工具
# 使用单条 RUN 指令减少镜像层数
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache dmidecode ca-certificates

# 2. 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/hardware-exporter /usr/local/bin/hardware-exporter

# 3. 暴露应用端口（根据你代码中的配置）
EXPOSE 30012

# 4. 启动程序
ENTRYPOINT ["/usr/local/bin/hardware-exporter"]