# ユーザーごとのPayer設定 - クイックスタートガイド

## 🚀 3ステップで設定完了

### ステップ1️⃣: 自分のユーザーIDを確認

Discordチャンネルで以下のコマンドを実行：

```
!whoami
```

**応答例**:
```
👤 あなたの情報
ユーザーID: 123456789012345678
ユーザー名: hoshi
表示名: 星 太郎

💡 この情報を使ってPayerを設定できます！
```

このユーザーIDをメモしてください。

### ステップ2️⃣: main.goを編集

`main.go`の約105行目にある`getPayerFromDiscordUser`関数を編集します。

#### 編集前（デフォルト）:

```go
switch userID {
case "123456789012345678": // 例: ユーザーAのID
	log.Printf("  ✅ [Payer] ユーザーID %s を検出 -> Payer: S", userID)
	return "S"
case "987654321098765432": // 例: ユーザーBのID
	log.Printf("  ✅ [Payer] ユーザーID %s を検出 -> Payer: T", userID)
	return "T"
case "111222333444555666": // 例: ユーザーCのID
	log.Printf("  ✅ [Payer] ユーザーID %s を検出 -> Payer: U", userID)
	return "U"
}
```

#### 編集後（実際のユーザーID）:

```go
switch userID {
case "234567890123456789": // 太郎さんの実際のID
	log.Printf("  ✅ [Payer] ユーザーID %s を検出 -> Payer: S", userID)
	return "S"
case "345678901234567890": // 花子さんの実際のID
	log.Printf("  ✅ [Payer] ユーザーID %s を検出 -> Payer: T", userID)
	return "T"
case "456789012345678901": // 次郎さんの実際のID
	log.Printf("  ✅ [Payer] ユーザーID %s を検出 -> Payer: U", userID)
	return "U"
}
```

**ポイント**:
- `case "実際のユーザーID":` の部分を書き換える
- `return "S"` の部分がDifyに送信されるpayer値
- コメント（`//`以降）はわかりやすい名前に変更

### ステップ3️⃣: Botを再起動

```bash
# Ctrl+CでBotを停止（実行中の場合）
# その後、再起動
go run main.go
```

または、ビルドして実行：

```bash
go build -o discord-bot main.go
./discord-bot
```

## ✅ テスト方法

### 1. ユーザー情報を再確認

```
!whoami
```

### 2. 画像をアップロードしてテスト

画像を添付して以下を送信：

```
!upload
```

### 3. ログで確認

ターミナルに以下のようなログが表示されればOK：

```log
📷 [STEP 1/8] onMessageCreate: メッセージ受信 - Content: !upload, Author: hoshi
🔄 [STEP 7/8] Difyワークフロー実行開始 - FileID: xxx
  👤 [Dify Workflow] 実行ユーザー: ID=234567890123456789, Username=hoshi
  👤 [Payer] ユーザー情報: ID=234567890123456789, Username=hoshi
  ✅ [Payer] ユーザーID 234567890123456789 を検出 -> Payer: S
  💳 [Dify Workflow] 決定されたPayer: S
  📋 [Dify Workflow] inputs構造:
    - receipt_images: 配列（要素数: 1）
    - payer: S (ユーザー: hoshi)  ← ここを確認！
```

## 📋 設定例テンプレート

### 家族で使う場合

```go
switch userID {
case "111111111111111111": // パパ
	return "S"
case "222222222222222222": // ママ
	return "T"
case "333333333333333333": // 子供
	return "U"
}
```

### チームで使う場合

```go
switch userID {
case "444444444444444444": // 営業部 田中
	return "SALES_TANAKA"
case "555555555555555555": // 営業部 佐藤
	return "SALES_SATO"
case "666666666666666666": // 開発部 鈴木
	return "DEV_SUZUKI"
}
```

### グループでまとめる場合

```go
// 管理者グループ
adminIDs := []string{"777777777777777777", "888888888888888888"}
for _, adminID := range adminIDs {
	if userID == adminID {
		return "ADMIN"
	}
}

// 一般ユーザー
switch userID {
case "999999999999999999":
	return "USER_A"
default:
	return "GUEST" // 未登録ユーザー
}
```

## 🔍 デバッグ方法

### 問題: Payerが常にデフォルト値（S）になる

**原因**: ユーザーIDが一致していない

**確認方法**:
1. `!whoami`で表示されるユーザーIDをコピー
2. `main.go`の`case`文と完全一致するか確認
3. 半角・全角、スペースの有無をチェック

**ログの見方**:
```log
👤 [Payer] ユーザー情報: ID=234567890123456789, Username=hoshi
⚠️  [Payer] 未登録ユーザー（ID: 234567890123456789, Username: hoshi） -> デフォルトPayer: S
```
→ この場合、`234567890123456789`が`case`に登録されていない

### 問題: Botが起動しない

**原因**: 構文エラー

**確認方法**:
```bash
go build -o discord-bot main.go
```

エラーメッセージを確認して修正。

### 問題: ユーザーIDが取得できない

**確認**:
- Botに`MESSAGE CONTENT INTENT`が有効になっているか確認
- Discord Developer Portalで設定

## 💡 Tips

### Tip 1: 複数のPayerを同じユーザーに割り当てる

```go
case "123456789012345678":
	// 時間帯で切り替え
	hour := time.Now().Hour()
	if hour < 12 {
		return "S_MORNING"
	}
	return "S_AFTERNOON"
```

### Tip 2: ユーザー名でフォールバック

ユーザーIDの登録を忘れた場合のために、ユーザー名でも判定できます：

```go
// ユーザーIDで判定できなかった場合
switch username {
case "hoshi":
	return "S"
case "alice":
	return "T"
}
```

### Tip 3: 環境変数で管理（上級者向け）

`.env`ファイルに追加：

```env
USER_123456789012345678=S
USER_234567890123456789=T
USER_345678901234567890=U
```

`main.go`で読み込み：

```go
func getPayerFromDiscordUser(userID, username string) string {
	envKey := fmt.Sprintf("USER_%s", userID)
	if payer := os.Getenv(envKey); payer != "" {
		return payer
	}
	return "S" // デフォルト
}
```

**メリット**: コードを変更せずに設定を更新できる

## 🎓 よくある質問

**Q: ユーザーIDは変わることがある？**  
A: いいえ、DiscordユーザーIDは永続的で変わりません。

**Q: ユーザー名は使わない方がいい？**  
A: はい。ユーザー名は変更可能なので、ユーザーIDを推奨します。

**Q: デフォルト値を変更したい**  
A: 関数の最後の`return "S"`を変更してください。

```go
// デフォルト値
log.Printf("  ⚠️  [Payer] 未登録ユーザー -> デフォルトPayer: UNKNOWN")
return "UNKNOWN"
```

**Q: case文の数に制限はある？**  
A: いいえ、必要なだけ追加できます。

```go
case "ID1":
	return "PAYER1"
case "ID2":
	return "PAYER2"
// ... 100個でも1000個でもOK
```

## 📞 サポート

問題が発生した場合は、以下の情報を添えて質問してください：

1. `!whoami`の出力結果
2. `main.go`の`getPayerFromDiscordUser`関数のコード
3. Botのログ（特に`[Payer]`部分）
4. エラーメッセージ（あれば）

## 🎉 完了！

これでユーザーごとに異なるPayerを設定できるようになりました！

次は実際に`!upload`コマンドで画像をアップロードして、Difyワークフローが正しいPayer値を受け取っているか確認してください。
