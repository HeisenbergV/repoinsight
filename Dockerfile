# 使用多阶段构建
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/repoinsight .

# 最终镜像
FROM alpine:latest

# 安装基本工具
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN adduser -D -g '' repoinsight

# 创建必要的目录
RUN mkdir -p /app/logs && chown -R repoinsight:repoinsight /app

# 切换到非 root 用户
USER repoinsight

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/repoinsight .
# 复制配置文件
COPY --from=builder /app/config.yml /app/config.yml

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./repoinsight"] 