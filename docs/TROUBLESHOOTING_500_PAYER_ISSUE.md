# Difyワークフロー400/500エラー（Payer形式）トラブルシューティング

## 📋 事象の概要

### 発生した問題（解決済み）

#### エラー1: HTTP 400 Bad Request（当初の根本原因）
- **エラーコード**: HTTP 400 Bad Request
- **エラーメッセージ**: 
  ```json
  {"code":"invalid_param","message":"payer in input form must be one of the following: ['\"Y\"', '\"S\"']","status":400}
  ```
- **原因**: Difyワークフローが期待するpayer値の形式が **`"\"Y\""`** または **`"\"S\""`**（エスケープされた二重引用符付き）だった
- **解決方法**: Dify側の設定を `Y` と `S`（シンプルな文字列）に変更

#### エラー2: HTTP 500 Internal Server Error（誤った修正後）
- **エラーコード**: HTTP 500 Internal Server Error
- **エラーメッセージ**: 
  ```json
  {"code":"internal_server_error","message":"The server encountered an internal error and was unable to complete your request.","status":500}
  ```
- **発生タイミング**: パートナー（未登録ユーザー）が画像をアップロードした時
- **正常動作**: 登録済みユーザー（Payer "Y"）は正常に動作していた

### 環境情報
- **Bot**: Discord Bot (budget-book-discord-bot)
- **API**: Dify Workflow API
- **対象チャンネル**: hoge

## 🔍 事象の詳細

### 成功パターン
| ユーザー | Discord UserID | Username | 判定されたPayer | 結果 |
|---------|----------------|----------|----------------|------|
| あなた | foo | Y | Y | ✅ 成功 |

### 失敗パターン
| ユーザー | Discord UserID | Username | 判定されたPayer | 結果 |
|---------|----------------|----------|----------------|------|
| パートナー | (未登録) | (未登録) | S (デフォルト) | ❌ 500エラー |

## 🔬 原因分析と解決

### ✅ 当初の原因: Difyワークフローが複雑な形式を期待していた

**当初のDifyワークフローは、payer値を以下の複雑な形式で期待していました:**

```json
{
  "payer": "\"Y\""  // または "\"S\""
}
```

つまり、**エスケープされた二重引用符付き文字列**が必要でした。

### 🎯 最終的な解決方法

**Dify側の入力変数定義を簡素化:**
- 変更前: `"Y"` と `"S"`（ダブルクォート付き文字列）
- 変更後: `Y` と `S`（シンプルな文字列）

これにより、Go側のコードも非常にシンプルになりました。

### なぜ "Y" は動作していたのか

#### 初期のコード（動作していた）
```go
"payer": fmt.Sprintf(`"%s"`, payer)
```

この実装により:
- `payer = "Y"` の場合 → `fmt.Sprintf(\`"%s"\`, "Y")` → `"Y"` （文字列）
- `json.Marshal()` によって → `"\"Y\""` （JSONエンコード後）
- Difyが期待する形式と一致 ✅

#### 誤った修正後（動作しなかった）
```go
"payer": payer  // 直接代入
```

この実装により:
- `payer = "S"` の場合 → `"S"` （文字列）
- `json.Marshal()` によって → `"S"` （JSONエンコード後）
- Difyが期待する形式 `"\"S\""` と不一致 ❌

### 混乱の原因

1. **Difyのエラーメッセージが不明瞭だった**
   - 初期は500エラー（内部エラー）で原因が分からなかった
   - 400エラーで初めて期待される形式が判明: `['\"Y\"', '\"S\"']`

2. **json.Marshal()の挙動に対する誤解**
   - `fmt.Sprintf(\`"%s"\`, payer)` は不要だと思っていた
   - しかし、Difyが特殊な形式を期待していたため、この処理が**必須**だった

## 🔧 実施した対処方法

### 1. デバッグログの強化

#### Payer判定のログ追加
```go
// DiscordユーザーからPayerを判定
payer := getPayerFromDiscordUser(userID, username)
log.Printf("🔑 判定されたPayer: %s (UserID: %s, Username: %s)", payer, userID, username)
```

#### 送信JSONのログ追加
```go
// デバッグ用: 送信するJSONをログ出力
log.Printf("📤 Difyへ送信するJSON: %s", string(jsonData))
```

#### エラー時の詳細ログ追加
```go
if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
    log.Printf("❌ ワークフロー実行失敗 - ステータス: %d, UserID: %s, Payer: %s", 
        resp.StatusCode, userID, getPayerFromDiscordUser(userID, username))
    log.Printf("📥 Difyからのエラーレスポンス: %s", string(respBody))
    
    // 500エラーの場合はDifyサーバー側の問題を指摘
    if resp.StatusCode == 500 {
        log.Printf("⚠️  Difyサーバー内部エラー: ワークフロー内のロジックやプラグインを確認してください")
    }
}
```

### 2. !whoamiコマンドの改善

ユーザーが自身のPayer判定結果を確認できるように改善:

```go
// !whoamiコマンド
if m.Content == "!whoami" {
    // 現在のPayer判定結果も表示
    currentPayer := getPayerFromDiscordUser(m.Author.ID, m.Author.Username)
    userInfo := fmt.Sprintf("👤 **あなたの情報**\n```\nユーザーID: %s\nユーザー名: %s\n表示名: %s\n現在のPayer: %s\n```\n💡 この情報を使ってPayerを設定できます！",
        m.Author.ID, m.Author.Username, m.Author.GlobalName, currentPayer)
    _, _ = s.ChannelMessageSend(m.ChannelID, userInfo)
    
    // ログにも出力
    log.Printf("📋 !whoami実行 - UserID: %s, Username: %s, Payer: %s", 
        m.Author.ID, m.Author.Username, currentPayer)
    return
}
```

### 3. シンプルなpayer形式の実装（最終版）

Dify側の設定変更により、コードが非常にシンプルになりました:

```go
// DiscordユーザーからPayerを判定
payer := getPayerFromDiscordUser(userID, username)
log.Printf("🔑 判定されたPayer: %s (UserID: %s, Username: %s)", payer, userID, username)

requestBody := map[string]interface{}{
    "inputs": map[string]interface{}{
        difyInputName: []interface{}{imageData}, // 配列形式で送信
        "payer":       payer,                    // "Y" または "S" を直接送信
    },
    "response_mode": "blocking",
    "user":          "discord-bot-user",
}
```

#### 重要ポイント
- **複雑な文字列変換は不要**になりました
- `payer` を直接送信するだけでOK
- Dify側で `Y` と `S` をシンプルな文字列として受け付けるように変更したため

## 📊 検証手順

### ステップ1: ユーザー情報の確認
```
1. パートナーにDiscordで `!whoami` コマンドを実行してもらう
2. 出力された情報を記録:
   - ユーザーID
   - ユーザー名
   - 現在のPayer
```

### ステップ2: ログの確認
```
Botを再起動後、以下のログを確認:

1. 🔑 判定されたPayer: <値>
2. 📤 Difyへ送信するJSON: {...}
3. ❌ ワークフロー実行失敗の場合: 
   - ステータスコード
   - 📥 Difyからのエラーレスポンス
```

### ステップ3: Dify側の確認
```
Dify管理画面で以下を確認:

1. ワークフロー実行ログ
   - どのノードでエラーが発生しているか
   - Payer "S" の場合の処理フロー

2. データベース/API設定
   - Payer "S" 用のデータが正しく設定されているか
   - 必要な環境変数や認証情報が設定されているか

3. プラグイン設定
   - Payer値による条件分岐が正しく設定されているか
```

### ステップ4: 再現テスト
```
1. パートナーに別の画像で再度試してもらう
2. あなた（Payer "Y"）でも同じ画像を試す
3. パートナーのUserIDを登録して Payer "Y" として試す
```

## 🎯 根本解決のための推奨アクション

### 即座に実施すべきこと

#### 1. パートナーのUserID登録
```go
// main.go の getPayerFromDiscordUser 関数
func getPayerFromDiscordUser(userID, username string) string {
    // ユーザーIDで判定（優先）
    switch userID {
    case "123456789012345678": // 例: ユーザーAのID
        return "S"
    case "796223697559748648": // あなたのID
        return "Y"
    case "PARTNER_USER_ID_HERE": // ← パートナーのIDを追加
        return "S" // または "Y"
    }
    // ...
}
```

#### 2. Difyワークフローの確認
```
□ Payer "S" の処理フローを手動でテスト
□ データベースに "S" 用のサンプルデータを投入
□ エラーが発生しているノードを特定
□ プラグインのログを確認
```

### 中長期的な改善

#### 1. エラーハンドリングの強化
```go
// Dify側のエラーメッセージをより詳細に解析
if errorData, hasError := workflowResp["error"]; hasError {
    log.Printf("⚠️  Dify内部エラーを検出: %v", errorData)
    
    // エラー種別ごとの処理
    if strings.Contains(fmt.Sprintf("%v", errorData), "PluginDaemonInnerError") {
        log.Printf("プラグインエラー: 管理画面でワークフローのログを確認")
    }
}
```

#### 2. リトライ機構の実装
```go
// 500エラーの場合は1回だけリトライ
if resp.StatusCode == 500 {
    log.Printf("⚠️  500エラー検出 - 3秒後にリトライします...")
    time.Sleep(3 * time.Second)
    // リトライ処理
}
```

#### 3. フォールバック処理
```go
// Difyがエラーの場合は別の処理パスを用意
if difyFailed {
    log.Printf("Difyエラー - ローカル処理にフォールバック")
    // 簡易的な処理を実行
}
```

## 📚 関連ドキュメント

- [USER_PAYER_MAPPING.md](./USER_PAYER_MAPPING.md) - Payer設定ガイド
- [TROUBLESHOOTING_502.md](./TROUBLESHOOTING_502.md) - 502エラー対応
- [DEBUG_401_IMPLEMENTATION.md](./DEBUG_401_IMPLEMENTATION.md) - 認証エラー対応
- [DEVELOPMENT_GUIDE.md](./DEVELOPMENT_GUIDE.md) - 開発ガイド

## 🔗 参考情報

### Payer判定ロジック
```go
// 優先順位:
// 1. Discord UserID（最優先）
// 2. Discord Username（フォールバック）
// 3. デフォルト値 "S"（未登録の場合）
```

### Dify APIエンドポイント
```
POST {DIFY_ENDPOINT}/workflows/run

リクエストボディ（最終版 - シンプルな形式）:
{
  "inputs": {
    "receipt_images": [{
      "transfer_method": "local_file",
      "upload_file_id": "<file_id>",
      "type": "image"
    }],
    "payer": "S" または "Y"  ← シンプルな文字列（Dify側で設定変更済み）
  },
  "response_mode": "blocking",
  "user": "discord-bot-user"
}

変更履歴:
- 旧形式: "payer": "\"S\""（複雑なエスケープ形式）
- 新形式: "payer": "S"（シンプルな文字列） ✅
```

## 📝 更新履歴

| 日付 | 内容 | 担当 |
|------|------|------|
| 2025-11-21 | 初版作成 - 500エラー調査と対処方法をドキュメント化 | GitHub Copilot |
| 2025-11-21 | Dify側の設定変更に伴い、シンプルな実装に更新 | GitHub Copilot |

---

## 💡 Tips

### デバッグ時のチェックリスト
- [ ] `!whoami` でUserIDとPayer判定を確認
- [ ] ログで送信JSONの内容を確認
- [ ] Dify管理画面でワークフロー実行ログを確認
- [ ] 別の画像で再現するか確認
- [ ] 登録済みユーザーと未登録ユーザーで比較テスト

### よくある誤解（当初の問題）
❌ デフォルト値 "S" が原因でエラーになる
→ ⭕ デフォルト値は正常に機能。問題はpayer値の**形式**だった

❌ 二重引用符の処理（`fmt.Sprintf(\`"%s"\`, payer)`）は不要
→ ⭕ **当初は**Difyワークフローが特殊な形式を期待していたため必須だった（現在は不要）

❌ `json.Marshal()` だけで正しい形式になる
→ ⭕ **当初は**事前に `"Y"` という文字列を作る必要があった（現在は不要）

❌ UserIDの登録が必須
→ ⭕ デフォルト値があるため必須ではないが、登録推奨

### 学んだこと
1. **エラーメッセージを注意深く読む**: 400エラーの `must be one of the following: ['\"Y\"', '\"S\"']` が決定的なヒントだった
2. **Difyの入力変数定義を確認**: ワークフロー側で期待される形式を把握することが重要
3. **デバッグログの重要性**: 送信するJSON全体をログ出力することで問題を早期発見できる
4. **複雑すぎる設定は見直す**: Dify側の設定をシンプルにすることで、コードも保守性も向上した ✨
