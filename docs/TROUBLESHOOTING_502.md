# Dify 502 Bad Gateway エラー トラブルシューティング

## 問題: PluginDaemonInnerError

### エラーメッセージ
```
req_id: ec64f5b62b PluginInvokeError: {
  "args": null,
  "error_type": "PluginDaemonInnerError",
  "message": "encountered an error: invalid character '<' looking for beginning of value status: 502 Bad Gateway original response: <html>"
}
```

このエラーは、**Difyワークフロー内部で使用しているプラグインやAPIが502エラーを返している**ことを示しています。

## 原因

### 1. 外部APIの問題
- OpenAI API、Claude API、その他のLLM APIが応答しない
- APIキーの有効期限切れまたは無効
- レート制限（Rate Limit）に達している
- APIサービスのダウンタイム

### 2. Difyプラグインの問題
- プラグインの設定が不正
- プラグインのタイムアウト
- プラグイン間の依存関係の問題

### 3. ネットワークの問題
- Difyサーバーから外部APIへの接続がブロックされている
- タイムアウト設定が短すぎる
- ファイアウォールやプロキシの問題

### 4. Difyサービスの問題
- Difyのサービスが一時的にダウンしている
- 内部サーバーエラー
- メンテナンス中

## 診断方法

### 1. Difyワークフローのログを確認

1. [Dify管理画面](https://cloud.dify.ai/)にログイン
2. 対象のワークフローを開く
3. **Logs**（ログ）タブをクリック
4. 最新の実行ログを確認
5. どのノードでエラーが発生したかを特定

**確認するポイント**:
- エラーが発生したノード名
- エラーメッセージの詳細
- 各ノードの実行時間
- タイムアウトの有無

### 2. ワークフロー内のノードを個別にテスト

1. ワークフローの各ノードを一つずつ無効化
2. 最小構成でテスト実行
3. どのノードが502エラーの原因か特定

**テスト順序**:
```
Start → [問題のノード] → End
```

最小構成でテストして、問題のないノードを徐々に追加していく。

### 3. 外部API接続の確認

#### OpenAI API を使用している場合

**Dify管理画面で確認**:
1. **Settings** → **Model Provider**
2. OpenAI の設定を確認
3. API Keyが有効か確認
4. **Test Connection**をクリック

**API Keyのテスト（ターミナル）**:
```bash
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer YOUR_API_KEY"
```

正常な場合: モデル一覧のJSONが返る
エラーの場合: 401, 429, 502などのエラー

#### Claude API を使用している場合

**Anthropic API のテスト**:
```bash
curl https://api.anthropic.com/v1/messages \
  -H "x-api-key: YOUR_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{"model":"claude-3-sonnet-20240229","max_tokens":10,"messages":[{"role":"user","content":"Hi"}]}'
```

### 4. Botのログを詳細確認

Bot実行時のログで以下を確認:

```log
✅ [STEP 7/8 完了] Difyワークフロー実行成功 - ステータス: 200
⚠️  [Dify Workflow] Dify内部エラーを検出:
  エラー内容: {"error_type":"PluginDaemonInnerError",...}
🔍 [Dify Workflow] PluginDaemonInnerError診断:
  ❌ Difyワークフロー内のプラグインでエラーが発生しました
  原因の可能性:
  1. ワークフロー内のHTTPリクエストノードが502エラーを返している
  2. 外部APIへの接続がタイムアウトしている
  3. プラグインの設定が不正
  4. Difyサービスの一時的な問題
```

## 解決方法

### 方法1: API Keyの確認と再設定

1. Dify管理画面で使用しているAPI Keyを確認
2. 有効期限が切れていないか確認
3. 必要に応じて新しいAPI Keyを生成
4. Difyの**Model Provider**設定で更新
5. **Test Connection**で接続確認

### 方法2: タイムアウト設定の調整

ワークフロー内のHTTPリクエストノードやLLMノードで:
- **Timeout**設定を確認（デフォルト: 30秒）
- 必要に応じて60秒〜120秒に延長
- 特に画像処理を含む場合は長めに設定

### 方法3: レート制限の確認

**OpenAI の場合**:
- [OpenAI Usage Dashboard](https://platform.openai.com/usage)でクォータを確認
- レート制限に達している場合は、プランのアップグレードを検討
- または実行頻度を調整

**Claude の場合**:
- [Anthropic Console](https://console.anthropic.com/)で使用状況を確認

### 方法4: ワークフローの簡素化

複雑なワークフローの場合:
1. 不要なノードを削除
2. 並列処理を直列処理に変更
3. 複数のAPI呼び出しを減らす
4. キャッシュを有効化

### 方法5: 代替プロバイダーの使用

一時的な回避策として:
- OpenAI → Azure OpenAI
- Claude → GPT-4
- 別のLLMプロバイダーを試す

### 方法6: リトライ処理の実装

Bot側でリトライ機能を追加:

```go
maxRetries := 3
var result string
var err error

for i := 0; i < maxRetries; i++ {
    result, err = runDifyWorkflowWithImage(fileID)
    if err == nil {
        break
    }
    log.Printf("リトライ %d/%d: %v", i+1, maxRetries, err)
    time.Sleep(time.Second * 5) // 5秒待機
}
```

## チェックリスト

- [ ] Difyワークフローのログを確認した
- [ ] エラーが発生したノードを特定した
- [ ] 外部API（OpenAI/Claude等）のAPI Keyが有効
- [ ] API使用量がクォータ内
- [ ] タイムアウト設定が適切
- [ ] ネットワーク接続が正常
- [ ] Difyサービスが正常稼働中
- [ ] 最小構成でのテストを実施
- [ ] Bot側のログを確認

## よくあるケース

### ケース1: OpenAI API キーの問題

**症状**: 502エラーが継続的に発生

**解決方法**:
1. OpenAI Platform でAPI Keyを確認
2. 新しいAPI Keyを生成
3. Difyで更新
4. テスト実行

### ケース2: レート制限

**症状**: 短時間に複数回実行すると502エラー

**解決方法**:
1. API使用量を確認
2. 実行間隔を空ける
3. プランをアップグレード

### ケース3: タイムアウト

**症状**: 大きな画像や複雑な処理で502エラー

**解決方法**:
1. タイムアウトを延長（60秒→120秒）
2. 画像サイズを縮小
3. 処理を分割

### ケース4: Difyサービスの問題

**症状**: 全てのワークフローで502エラー

**解決方法**:
1. [Dify Status Page](https://status.dify.ai/)（存在する場合）を確認
2. Difyコミュニティで同様の報告がないか確認
3. 時間を置いて再試行

## 参考情報

### Difyのレスポンス構造

正常時:
```json
{
  "workflow_run_id": "abc123",
  "task_id": "def456",
  "data": {
    "outputs": {...}
  }
}
```

エラー時:
```json
{
  "error": {
    "error_type": "PluginDaemonInnerError",
    "message": "..."
  }
}
```

### サポート

- [Dify Documentation](https://docs.dify.ai/)
- [Dify GitHub Issues](https://github.com/langgenius/dify/issues)
- [Dify Discord Community](https://discord.gg/dify)

## 更新履歴

- 2025/11/02: 初版作成
