# Discord Bot with Dify Integration

Discord上で画像を受け取り、Difyに送信して処理するBotです。

## 📋 必要な準備

### 1. Go言語のインストール
- Go 1.24.5以上が必要です

### 2. Discord Bot の作成
1. [Discord Developer Portal](https://discord.com/developers/applications)にアクセス
2. 新しいApplicationを作成
3. Botタブから`APPLICATION_ID`, `PUBLIC_KEY`, `DISCORD_TOKEN`を取得
4. Bot Permissionsで以下を有効化：
   - `Send Messages`
   - `Read Messages/View Channels`
   - `Read Message History`
   - `Attach Files`
5. OAuth2 URLでBotをサーバーに招待

### 3. Dify API設定
1. [Dify](https://dify.ai/)にログイン
2. ワークフローまたはチャットボットアプリを作成
3. API Keyを取得
4. ワークフローで画像入力を受け取れるように設定

## 🔧 環境変数の設定

`.env`ファイルをプロジェクトルートに作成し、以下を設定してください：

```env
# Discord設定
APPLICATION_ID=your_application_id
PUBLIC_KEY=your_public_key
DISCORD_TOKEN=your_bot_token

# Dify設定
DIFY_API_KEY=app-xxxxxxxxxxxx
DIFY_ENDPOINT=https://api.dify.ai/v1
DIFY_WORKFLOW_ID=your_workflow_id  # オプション：ワークフロー使用時
```

## 🚀 起動方法

1. **プロジェクトディレクトリに移動**
```bash
cd /Users/hoshi/pg/discord-bot/1
```

2. **依存関係のインストール**
```bash
go mod download
```

3. **Botの起動**
```bash
go run main.go
```

起動すると以下のようなログが表示されます：
```
登録: /hello
登録: /hello2
登録: /hello3
Bot 起動中 Ctrl+Cで終了
```

## 📱 使用方法

### スラッシュコマンド

1. **`/hello`** - 簡単な挨拶を返します
2. **`/hello2`** - マルチセレクトメニューを表示（最大25個）
3. **`/hello3`** - ボタンを表示（最大5個）

### 画像処理機能

**`!upload` コマンド + 画像添付**

1. Discordチャンネルで画像を添付
2. メッセージに `!upload` と入力して送信
3. Botが画像を**自動的に圧縮**（最大1500px、品質85%）
4. 圧縮した画像をDifyに送信
5. Difyで処理した結果がDiscordに返ってきます

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
- **対応形式**: JPEG、PNG、GIF、BMP、WebP → JPEG変換

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

## 🔄 停止方法

`Ctrl+C` でBotを終了できます。

## 📂 プロジェクト構成

```
/Users/hoshi/pg/discord-bot/1/
├── main.go          # メインプログラム
├── go.mod           # Go依存関係管理
├── go.sum           # 依存関係チェックサム
├── .env             # 環境変数設定（要作成）
├── .gitignore       # Gitで無視するファイル
└── README.md        # このファイル
```

## 🔐 セキュリティ注意事項

- `.env`ファイルは絶対にGitにコミットしないでください
- Discord TokenやDify API Keyは第三者に漏らさないでください
- 本番環境では環境変数を適切に管理してください
