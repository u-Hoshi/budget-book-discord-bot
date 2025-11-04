# Koyeb Deep Sleep 問題対応ドキュメント

## 🚨 問題の概要

KoyebでDiscord Botをデプロイした際、以下のメッセージが表示されてインスタンスが停止する問題が発生しました。

```
No traffic detected in the past 300 seconds. Transitioning to deep sleep.
Instance is stopping.
```

## 📋 問題の詳細

### 発生条件
- **プラットフォーム**: Koyeb
- **アプリケーション**: Discord Bot (Go言語)
- **トリガー**: 300秒間（5分間）トラフィックが検出されない状態

### 根本原因
Koyebは無料プランにおいて、5分間リクエストが来ないとインスタンスを自動的にDeep Sleepモードに移行させる仕様があります。Discord Botのような常時稼働が必要なアプリケーションでは、この機能により予期しない停止が発生します。

## 🔧 解決方法

### 1. ヘルスチェック機能の実装

アプリケーション内で定期的に自分自身のヘルスチェックエンドポイントにリクエストを送信する機能を実装しました。

```go
// 5分間隔のヘルスチェック
func startHealthCheckCron() {
    // 環境変数からヘルスチェックURLを取得
    healthCheckURL := os.Getenv("HEALTH_CHECK_URL")
    if healthCheckURL == "" {
        port := os.Getenv("PORT")
        if port == "" {
            port = "8080"
        }
        healthCheckURL = fmt.Sprintf("http://localhost:%s", port)
    }
    
    // 5分間隔のティッカーを作成
    ticker := time.NewTicker(5 * time.Minute)
    
    go func() {
        defer ticker.Stop()
        for range ticker.C {
            performHealthCheck(healthCheckURL)
        }
    }()
}
```

### 2. 実装のポイント

#### 📊 間隔設定の検証結果
| 間隔 | 結果 | 備考 |
|------|------|------|
| 10分 | ❌ 失敗 | Deep Sleepが発動（5分制限） |
| 30分 | ❌ 失敗 | Deep Sleepが発動 |
| 5分 | ✅ 成功 | 制限時間内にリクエスト送信 |

#### 🎯 最適解
- **間隔**: 5分（300秒以内）
- **初回実行**: 起動5秒後
- **継続実行**: 5分間隔で自動継続

### 3. 環境変数設定

Koyebダッシュボードで以下の環境変数を設定：

```bash
HEALTH_CHECK_URL=https://your-app-name.koyeb.app
PORT=8080  # Koyebが自動設定
```

## 📈 効果測定

### Before（対策前）
```
🚀 Discord Bot 起動中...
✅ Bot起動完了 - Ctrl+Cで終了
（5分後）
No traffic detected in the past 300 seconds. Transitioning to deep sleep.
Instance is stopping.
```

### After（対策後）
```
🚀 Discord Bot 起動中...
🌐 HTTPサーバーを開始: ポート 8080
🕐 ヘルスチェックの定期実行を開始しました (5分間隔)
🔗 ヘルスチェックURL: https://your-app.koyeb.app
🔍 [2025-11-04 10:00:05] ヘルスチェック実行中...
✅ [2025-11-04 10:00:05] ヘルスチェック成功: 200
✅ Bot起動完了 - Ctrl+Cで終了
🔍 [2025-11-04 10:05:00] ヘルスチェック実行中...
✅ [2025-11-04 10:05:00] ヘルスチェック成功: 200
（継続稼働）
```

## ⚠️ 注意事項

### 1. リソース使用量
- 定期的なHTTPリクエストによりわずかなCPU・ネットワーク使用量が発生
- 5分間隔なので影響は最小限

### 2. 他プラットフォームでの動作
- **Heroku**: 30分制限（より緩い）
- **Railway**: 制限なし
- **Google Cloud Run**: リクエストベース課金

### 3. トラブルシューティング
Deep Sleepが依然として発生する場合：
1. `HEALTH_CHECK_URL`が正しく設定されているか確認
2. アプリケーションのログでヘルスチェックが実行されているか確認
3. 間隔を4分に短縮することを検討

## 🚀 デプロイ手順

1. **コード修正**: ヘルスチェック機能を実装
2. **環境変数設定**: `HEALTH_CHECK_URL`を設定
3. **デプロイ実行**: 修正版をKoyebにデプロイ
4. **動作確認**: ログでヘルスチェックが実行されることを確認

## 📚 参考リンク

- [Koyeb Deep Sleep Documentation](https://www.koyeb.com/docs/concepts/services#deep-sleep)
- [Discord Bot 継続稼働ベストプラクティス](https://discord.com/developers/docs/topics/gateway)

---

**最終更新**: 2025年11月4日  
**対象バージョン**: Go 1.24.5, Koyeb Free Plan