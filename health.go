package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// ヘルスチェック用のレスポンス構造体
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
	Uptime    string    `json:"uptime,omitempty"`
}

var startTime = time.Now()

// 定期的なヘルスチェック機能
func StartHealthCheckCron() {
	// 環境変数からヘルスチェックURLを取得
	healthCheckURL := os.Getenv("HEALTH_CHECK_URL")
	if healthCheckURL == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		healthCheckURL = fmt.Sprintf("http://localhost:%s", port)
	}

	log.Printf("🕐 ヘルスチェックの定期実行を開始しました (5分間隔)")
	log.Printf("🔗 ヘルスチェックURL: %s", healthCheckURL)

	// 初回ヘルスチェック（5秒後に実行）
	go func() {
		time.Sleep(5 * time.Second)
		performHealthCheck(healthCheckURL)
	}()

	// 5分間隔のティッカーを作成
	ticker := time.NewTicker(5 * time.Minute)

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			performHealthCheck(healthCheckURL)
		}
	}()
}

// ヘルスチェックを実行する関数
func performHealthCheck(url string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("🔍 [%s] ヘルスチェック実行中... (%s)", now, url)

	// タイムアウト付きのHTTPクライアント
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("❌ [%s] ヘルスチェックエラー: %v", now, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("✅ [%s] ヘルスチェック成功: %d", now, resp.StatusCode)
	} else {
		log.Printf("⚠️ [%s] ヘルスチェック失敗: %d", now, resp.StatusCode)
	}
} // 文字列を指定した長さに切り詰める

// ヘルスチェックエンドポイント
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)

	health := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    uptime.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}
