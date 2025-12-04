# Budget Book Discord Bot

**レシート画像を撮ってDiscordに投稿するだけで、自動的に家計簿に記録される**

Discord Bot × AI（Dify） × Google Spreadsheet を組み合わせた、完全自動化家計簿システムです。

[![Go](https://img.shields.io/badge/Go-1.24.5-00ADD8?logo=go)](https://go.dev/)
[![Discord](https://img.shields.io/badge/Discord-Bot-5865F2?logo=discord&logoColor=white)](https://discord.com/)
[![Dify](https://img.shields.io/badge/Dify-AI-FF6B00)](https://dify.ai/)

## 📖 プロジェクト概要

レシート画像をスマホで撮影してDiscordに投稿するだけで、AI（Dify）が自動的に以下の情報を抽出し、Google Spreadsheetに記録します：

- 🏪 **店舗名**
- 💰 **金額**
- 📝 **購入項目**（食費、日用品など）

家計簿アプリへの手入力が不要になり、継続的な記録が可能になります。

## 🎥 デモ・発表資料

このプロジェクトについて、コミュニティで発表を行いました。

[![スライドタイトル](/images/slide-title.png)](https://speakerdeck.com/u_hoshi/dify-xspreadsheetsdezuo-rujia-ji-bo-turu)

📊 [発表スライド](https://speakerdeck.com/u_hoshi/dify-xspreadsheetsdezuo-rujia-ji-bo-turu)



## 🏗️ システム構成

[![システム構成図](/images/system-architecture.png)](https://speakerdeck.com/u_hoshi/dify-xspreadsheetsdezuo-rujia-ji-bo-turu?slide=12)

## ✨ 主な機能

### 🤖 自動レシート処理
- レシート画像を投稿するだけで自動処理
- 画像圧縮機能により、通信量を削減
- AI による高精度な文字認識

### 💬 Discord コマンド
- `いくら` - 今月の支出サマリーを表示
- `!whoami` - ユーザー情報確認
- `!ping` - Bot の動作確認

### 👥 複数人対応
- ユーザーごとの支出を自動で振り分け
- カップル・家族での共同管理に対応

### 🔄 ヘルスチェック機能
- 定期的な死活監視
- 無料ホスティングサービスのスリープ対策

## 🛠️ 技術スタック

| カテゴリ | 技術 |
|---------|-----|
| **言語** | Go 1.24.5 |
| **Bot Framework** | discordgo |
| **AI/画像解析** | Dify (gemini) |
| **データ保存** | Google Spreadsheet |
| **中間処理** | Google Apps Script |
| **ホスティング** | Koyeb (or Docker) |
| **画像処理** | imaging (圧縮・リサイズ) |

## 📂 プロジェクト構成

```
budget-book-discord-bot/
├── main.go          # メイン処理、メッセージハンドラー
├── health.go        # ヘルスチェック機能
├── image.go         # 画像ダウンロード・圧縮
├── dify.go          # Dify API連携
├── utils.go         # ユーティリティ関数
├── go.mod           # 依存関係管理
├── Dockerfile       # Docker設定
├── README.md        # このファイル（プロジェクト概要）
└── docs/
    └── README.md    # 開発者向け詳細ドキュメント
```

## 🚀 クイックスタート

```bash
# リポジトリのクローン
git clone https://github.com/u-Hoshi/budget-book-discord-bot.git
cd budget-book-discord-bot

# 環境変数の設定（.envファイルを作成）
# 必要な環境変数: DISCORD_TOKEN, DIFY_API_KEY, GAS_ENDPOINT など

# 依存関係のインストール
go mod tidy

# 起動
go run main.go
```

**詳細な開発手順・テスト方法は [docs/README.md](docs/README.md) を参照してください。**

## 📚 ドキュメント

- **[開発ガイド](docs/README.md)** - 開発環境構築、詳細な使い方
- [ユーザーマッピング設定](docs/USER_PAYER_MAPPING.md) - 複数人での支出管理
- [画像圧縮設定](docs/IMAGE_COMPRESSION.md) - 画像処理のカスタマイズ
- [トラブルシューティング](docs/TROUBLESHOOTING_401.md) - よくある問題と解決方法

## 💡 開発の背景

家計簿アプリへの手入力が面倒で継続できない問題を解決するために開発しました。

**課題:**
- レシートの手入力が面倒
- 家計簿アプリを開くのが億劫
- 継続できない

**解決策:**
- 普段使っているDiscordに投稿するだけ
- AI が自動で内容を読み取り
- Google Spreadsheet に自動記録

結果として、**家計簿の記録を継続できるようになりました**。


## 📝 ライセンス

このプロジェクトは個人利用目的で開発されました。

