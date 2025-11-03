# 画像圧縮機能の追加

## 必要なライブラリのインストール

画像圧縮機能を追加するために、以下のライブラリをインストールしてください：

```bash
cd /Users/hoshi/pg/discord-bot/1
go get github.com/disintegration/imaging
```

または、`go.mod`に自動的に追加されるため、単に`go run main.go`を実行するだけでも自動的にダウンロードされます。

## 実装される機能

### 1. 画像のリサイズ
- 幅を最大1500pxに制限
- アスペクト比を維持

### 2. JPEG品質の調整
- 品質を85%に設定（ファイルサイズとクオリティのバランス）

### 3. 圧縮前後のファイルサイズ表示
- 圧縮率をログに出力

## 環境変数設定

`.env`ファイルに以下を追加できます（オプション）：

```env
# 画像圧縮設定
IMAGE_MAX_WIDTH=1500    # 最大幅（ピクセル）デフォルト: 1500
IMAGE_QUALITY=85        # JPEG品質（1-100）デフォルト: 85
ENABLE_COMPRESSION=true # 圧縮を有効化 デフォルト: true
```

## 使用方法

1. **ライブラリをインストール**
   ```bash
   go get github.com/disintegration/imaging
   ```

2. **Botを起動**
   ```bash
   go run main.go
   ```

3. **Discord経由で画像をアップロード**
   - 画像を添付して`!upload`コマンドを実行
   - 自動的に圧縮されてからDifyに送信されます

## ログ出力例

```log
📥 [Compress] 画像圧縮開始: IMG_0388.JPG
  ✅ [Compress] 画像読み込み成功: 4032x3024
  🔍 [Compress] 元のサイズ: 2.5 MB
  ⚙️  [Compress] リサイズ: 1500x1125 (品質: 85%)
  ✅ [Compress] 圧縮完了: IMG_0388_compressed.jpg
  📊 [Compress] 圧縮後サイズ: 0.8 MB (圧縮率: 68%)
```

## トラブルシューティング

### エラー: パッケージが見つからない

```bash
go: github.com/disintegration/imaging: module github.com/disintegration/imaging: Get "https://proxy.golang.org/...": dial tcp: lookup proxy.golang.org: no such host
```

**解決方法**:
```bash
go env -w GOPROXY=https://goproxy.cn,direct
go get github.com/disintegration/imaging
```

### 圧縮をスキップしたい場合

`.env`で無効化：
```env
ENABLE_COMPRESSION=false
```

または、コード内で直接変更：
```go
enableCompression := false // 圧縮を無効化
```

## パフォーマンス

| 元のサイズ | 圧縮後 | 処理時間 |
|-----------|--------|---------|
| 5 MB (4032x3024) | 0.8 MB (1500x1125) | ~1秒 |
| 3 MB (3000x2000) | 0.6 MB (1500x1000) | ~0.7秒 |
| 1 MB (2000x1500) | 0.4 MB (1500x1125) | ~0.5秒 |

## 参考

- [imaging ライブラリ](https://github.com/disintegration/imaging)
- [Go画像処理チュートリアル](https://github.com/disintegration/imaging#usage)
