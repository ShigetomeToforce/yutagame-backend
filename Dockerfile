# 💡 1. ビルド用の環境（マルチステージビルドで成果物を軽量化します）
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 依存関係（go.mod / go.sum）を先にコピーしてキャッシュを利かせます
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをすべてコピー
COPY . .

# Goバイナリをビルド（CGOを無効化して軽量・安全なスタンドアロンバイナリにします）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .


# 💡 2. 実行用の超軽量環境
FROM alpine:latest

WORKDIR /app

# タイムゾーンデータをインストール（parseTime=True でMySQLの時間をJSTで扱うため）
RUN apk --no-cache add tzdata && \
    cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
    echo "Asia/Tokyo" > /etc/timezone

# builderステージからビルド済みのバイナリだけをコピー
COPY --from=builder /app/main .

# 🚨 超重要：ホスト側とマウントする「storage」の受け皿フォルダをあらかじめコンテナ内にも作っておく
RUN mkdir -p /app/storage

# サーバー起動コマンド
CMD ["./main"]