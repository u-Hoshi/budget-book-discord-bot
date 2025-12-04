# 開発者向けドキュメント

このドキュメントは、開発者や1年後の自分がスムーズにローカル環境で開発を再開できるようにするためのガイドです。

> **📌 プロジェクト概要を知りたい方は [../README.md](../README.md) を参照してください。**

---

## 🚀 ローカル開発環境のセットアップ

### 1. 前提条件

以下がインストールされていることを確認してください：

```bash
# Go のバージョン確認
go version
# Go 1.24.5 以上が必要

# Git の確認
git --version
```

### 2. プロジェクトのクローン

```bash
git clone https://github.com/u-Hoshi/budget-book-discord-bot.git
cd budget-book-discord-bot
```

### 3. Discord Bot の作成
1. [Discord Developer Portal](https://discord.com/developers/applications)にアクセス
2. 新しいApplicationを作成
3. Botタブから`APPLICATION_ID`, `PUBLIC_KEY`, `DISCORD_TOKEN`を取得
4. Bot Permissionsで以下を有効化：
   - `Send Messages`
   - `Read Messages/View Channels`
   - `Read Message History`
   - `Attach Files`
5. OAuth2 URLでBotをサーバーに招待

### 4. Dify API設定
1. [Dify](https://dify.ai/)にログイン
2. ワークフローまたはチャットボットアプリを作成
3. API Keyを取得
4. ワークフローで画像入力を受け取れるように設定

### 5. 環境変数の設定

プロジェクトルートに`.env`ファイルを作成し、以下を設定してください：

```bash
# プロジェクトルートに.envファイルを作成
touch .env
```

```.env
# Discord設定
APPLICATION_ID=your_application_id
DISCORD_TOKEN=your_bot_token

# Dify設定
DIFY_API_KEY=app-xxxxxxxxxxxx
DIFY_ENDPOINT=https://api.dify.ai/v1

# GAS設定
GAS_ENDPOINT=https://script.google.com/macros/s/xxxxx/exec

# オプション: 画像圧縮設定
IMAGE_MAX_WIDTH=1500
IMAGE_QUALITY=85
ENABLE_COMPRESSION=true

# オプション: ヘルスチェック設定
PORT=8080
HEALTH_CHECK_URL=http://localhost:8080
```

### 6. 依存関係のインストール

```bash
go mod tidy
```

### 7. Bot の起動
```bash
# 方法1: 直接実行
go run main.go

# 方法2: ビルドしてから実行
go build -o bot
./bot
```

起動すると以下のようなログが表示されます：
```
🚀 Discord Bot 起動中...
✅ 必要な環境変数が設定されています。
登録: /hello
🌐 HTTPサーバーを開始: ポート 8080
🕐 ヘルスチェックの定期実行を開始しました (5分間隔)
✅ Bot起動完了 - Ctrl+Cで終了
```

### 8. 動作確認

Discordサーバーで以下のコマンドを試してください：

| コマンド | 期待される動作 |
|---------|--------------|
| `!ping` | "Pong!" が返ってくる |
| `!whoami` | 自分のユーザー情報が表示される |
| `いくら` | 今月の家計簿サマリーが表示される |
| 画像投稿 | レシート解析結果が返ってくる |

---

## 📱 使用方法

### コマンド一覧


### 画像処理機能


1. Discordチャンネルで画像を添付
2. Botが画像を**自動的に圧縮**（最大1500px、品質85%）
3. 圧縮した画像をDifyに送信
4. Difyで処理した結果がDiscordに返ってきます

#### 処理フロー
```
Discord画像添付 → Bot受信 → ローカル一時保存 
→ 画像圧縮（リサイズ + 品質調整）
→ Difyファイルアップロード → Difyワークフロー実行 
→ 結果をDiscordに返信 → 一時ファイル削除
```

#### 画像圧縮の詳細
- **最大幅**: 1500px（アスペクト比維持）
- **JPEG品質**: 85%（高品質を維持しつつファイルサイズを削減）
- **圧縮率**: 通常60-80%のファイルサイズ削減

設定をカスタマイズするには、`.env`ファイルに以下を追加：
```env
IMAGE_MAX_WIDTH=1500        # 最大幅（px）
IMAGE_QUALITY=85            # JPEG品質（1-100）
ENABLE_COMPRESSION=true     # 圧縮ON/OFF
```

#### ユーザーごとのPayer切り替え

Difyワークフローに渡す`payer`の値は、メッセージを送信したDiscordユーザーに応じて自動的に切り替わります。

**設定方法**:
1. `!whoami`コマンドで自分のユーザーIDを確認
2. `main.go`の`getPayerFromDiscordUser`関数を編集
3. ユーザーIDとPayerのマッピングを設定

詳細は `docs/USER_PAYER_MAPPING.md` を参照してください。

### その他のコマンド

- **`!ping`** - Botの応答確認（"Pong!"を返します）
- **`!whoami`** - 自分のDiscordユーザー情報を表示（ID、ユーザー名、表示名）

## 🛠️ Dify側の設定

### ワークフロー設定例

1. **Start Node**で画像入力を受け取る設定
   - Input変数: `image` (type: `File`)
   
2. **処理ノード**で画像を分析
   - LLMノードやVisionモデルを使用
   
3. **End Node**で結果を返す

### inputs設定
ワークフローのinputsには以下の形式で画像が渡されます：
```json
{
  "image": {
    "transfer_method": "local_file",
    "upload_file_id": "ファイルID",
    "type": "image"
  }
}
```

## 🐛 トラブルシューティング

### Botが起動しない
- `.env`ファイルが正しく設定されているか確認
- `APPLICATION_ID`と`DISCORD_TOKEN`が正しいか確認

### 画像がアップロードできない
- `DIFY_API_KEY`が正しく設定されているか確認
- Difyのプランで画像アップロードが許可されているか確認
- ネットワーク接続を確認

### ワークフローが実行されない
- Dify側で画像入力を受け取る設定になっているか確認
- `DIFY_ENDPOINT`が正しいか確認（自己ホスト版の場合は独自のエンドポイント）

## 📝 ログ出力

Bot起動中は以下のようなログが出力されます：

```
📷 onMessageCreate Received message: !upload
📥 画像をダウンロードしました: IMG_0388.JPG
✅ Difyにファイルをアップロードしました: ID=xxx, Name=IMG_0388.JPG
✅ Difyワークフローを実行しました: {...}
```

---

## 🔧 開発時のコマンド集

### よく使うコマンド

```bash
# 依存関係の整理（新しいパッケージを追加した後など）
go mod tidy

# ビルドキャッシュのクリア（ビルドエラーが解決しない時）
go clean

# ビルド
go build -o bot

# すべてのGoファイルを確認
ls *.go
```

### テストコマンド

```bash
# 全テストを実行
go test -v

# カバレッジ付きで実行
go test -cover

# カバレッジの詳細を確認
go test -cover -coverprofile=coverage.out
go tool cover -func=coverage.out

# 特定のテストのみ実行
go test -v -run TestTruncate
go test -v -run TestGetPayer

# ベンチマークを実行
go test -bench=.
```

### トラブルシューティング

#### `undefined: XXX` エラーが出る

ビルドキャッシュの問題の可能性があります：

```bash
go clean
go mod tidy
go build
```

#### `.env` ファイルが読み込まれない

- `.env` ファイルが`main.go`と同じディレクトリにあるか確認
- ファイル名が正しく `.env`（ドットで始まる）になっているか確認
- 環境変数の値にスペースや改行が入っていないか確認

#### Bot がメッセージに反応しない

1. Discord Developer Portal で **Message Content Intent** が有効か確認
2. `DISCORD_TOKEN` が正しいか確認
3. Bot がサーバーに招待されているか確認
4. Bot に適切な権限（メッセージ送信・読み取り）があるか確認

#### Dify API でエラーが出る

- `DIFY_API_KEY` が正しいか確認
- Dify のワークフローが正しく設定されているか確認
- API の利用制限に達していないか確認

## 📂 プロジェクト構成

```
budget-book-discord-bot/
├── main.go          # メイン処理、メッセージハンドラー
├── health.go        # ヘルスチェック機能
├── image.go         # 画像ダウンロード・圧縮
├── dify.go          # Dify API連携
├── utils.go         # ユーティリティ関数
├── go.mod           # Go依存関係管理
├── go.sum           # 依存関係チェックサム
├── .env             # 環境変数（要作成・gitignore対象）
├── .gitignore       # Git管理外ファイル
├── Dockerfile       # Docker設定
├── README.md        # プロジェクト概要（外部向け）
└── docs/
    ├── README.md    # このファイル（開発者向け）
    └── ...          # その他のドキュメント
```

### ファイルの役割

| ファイル | 役割 |
|---------|------|
| `main.go` | エントリーポイント、Discord メッセージハンドラー |
| `health.go` | ヘルスチェック、HTTP サーバー |
| `image.go` | 画像のダウンロード・圧縮処理 |
| `dify.go` | Dify API との通信 |
| `utils.go` | 汎用的なユーティリティ関数 |

## 🔄 停止方法

`Ctrl+C` でBotをシャットダウンできます。

graceful shutdown が実装されているため、安全に終了できます。

## 🧪 テストについて

このプロジェクトでは、**実装のコアとなるビジネスロジック**にのみテストを記述しています。

### ✅ テスト対象（カバレッジ: 100%）

以下の純粋関数には、包括的なテストを実装しています：

| 関数 | 役割 | テストケース数 |
|------|------|---------------|
| `TruncateString()` | 文字列切り詰め（マルチバイト文字対応） | 6 |
| `FormatAmountWithComma()` | 金額のカンマ区切りフォーマット | 7 |
| `GetMimeType()` | ファイル拡張子からMIME type判定 | 8 |
| `getPayerFromDiscordUser()` | ユーザーID/名前からpayerを判定 | 7 |

### ❌ テスト不要と判断した部分

外部サービス（Dify、Google Spreadsheet、Discord）への橋渡しとなる関数は、**モックだらけのテストは価値が低い**ため、実装していません：

- `UploadImageToDify()` - Dify API呼び出し
- `RunDifyWorkflowWithImage()` - Difyワークフロー実行
- `DownloadImage()` / `CompressImage()` - 外部ライブラリのラッパー
- Discord関連のハンドラ - 外部サービス統合
- `main()` - 統合処理

### 🎯 テスト方針

処理の大部分をDifyやSpreadsheet側で行っているため、このBotは**「接着剤」の役割**です。

そのため、以下の方針でテストを実装しています：

1. **純粋関数のビジネスロジック** → テスト必須（100%カバレッジ）
2. **外部APIとの統合** → E2Eテストや手動テストで確認
3. **統合処理・フレームワーク呼び出し** → テスト不要

この方針により、**コストパフォーマンスの高いテスト**を実現しています。

### 🐛 テストで発見したバグ

テスト実装により、実際にバグを発見・修正しました

**問題**: `TruncateString()` が日本語（マルチバイト文字）を扱う際に文字が壊れる

```go
// 修正前（バグあり）
func TruncateString(s string, maxLen int) string {
    if len(s) <= maxLen {  // len(s) はバイト数
        return s
    }
    return s[:maxLen] + "...(省略)"  // バイト位置で切るため文字が壊れる
}

// 修正後
func TruncateString(s string, maxLen int) string {
    runes := []rune(s)  // 文字（ルーン）として扱う
    if len(runes) <= maxLen {
        return s
    }
    return string(runes[:maxLen]) + "...(省略)"
}
```

**影響**: Discordへの返信メッセージで日本語が文字化けする可能性があった  
**対応**: テストケース追加により今後同様のバグを防止

### 📊 カバレッジレポート

```bash
$ go test -cover
PASS
coverage: 7.6% of statements
```

全体カバレッジは7.6%ですが、**テスト対象関数は100%**です。これは、外部サービス連携のコードが大部分を占めるためです。

### 💡 テスト実行例

```bash
# 全テスト実行
$ go test -v
=== RUN   TestTruncateString
=== RUN   TestTruncateString/通常ケース
=== RUN   TestTruncateString/長い文字列
...
--- PASS: TestTruncateString (0.00s)
=== RUN   TestFormatAmountWithComma
...
PASS
ok      github.com/u-Hoshi/budget-book-discord-bot      0.204s
```

## 🔐 セキュリティ注意事項

- `.env`ファイルは絶対にGitにコミットしないでください
- Discord TokenやDify API Keyは第三者に漏らさないでください
- 本番環境では環境変数を適切に管理してください
