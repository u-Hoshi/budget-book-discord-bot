# Build stage
FROM golang:1.21-alpine AS builder

# セキュリティ更新とビルドに必要なパッケージをインストール
RUN apk update && apk add --no-cache git ca-certificates tzdata

# 作業ディレクトリを設定
WORKDIR /app

# Go modulesファイルをコピーして依存関係をダウンロード
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド（静的リンク、最適化）
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main .

# Runtime stage
FROM alpine:latest

# セキュリティ更新とCA証明書をインストール
RUN apk --no-cache add ca-certificates tzdata

# 非rootユーザーを作成
RUN adduser -D -s /bin/sh appuser

# 作業ディレクトリを設定
WORKDIR /app

# ビルドしたバイナリをコピー
COPY --from=builder /app/main .

# 実行権限を設定
RUN chmod +x main

# ユーザーを切り替え
USER appuser

# アプリケーションを実行
CMD ["./main"]
