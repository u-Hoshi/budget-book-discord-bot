了解です。以下は、今回の「**Mac上の画像サイズ表示（3.6MB）とコード上の送信サイズ（9.0MB）との差異**」に関する補足ドキュメント（`image_size_difference.md`）です。
リポジトリ内の `/docs/troubleshooting/` 配下などに配置することを想定しています。

---

````markdown
# 画像サイズの差異に関する補足ドキュメント

## 概要

MacのFinderで確認したファイルサイズ（例: **3.6MB**）と、  
コード上で送信時に確認されるサイズ（例: **9.0MB**）の間に差が発生した事象について、  
原因と対応を整理します。

---

## 現象

- Finder上での表示：3.6 MB  
- コード上（`len(base64.b64encode(image_bytes))`など）で確認：9.0 MB  
- API側では「ファイルサイズが上限を超過している」とエラーが返却される。

---

## 原因

### 1. Base64エンコードによるデータ膨張

画像をAPIリクエストに含める際、`multipart/form-data` や `JSON` 形式にするために  
**Base64エンコード**を行うケースがあります。

Base64エンコードは、  
バイナリデータをASCII文字列として安全に転送するための方式ですが、  
**元データより約33%大きくなります**。

例:

| 種別 | 内容 |
|------|------|
| 元データ | PNG, JPEGなどのバイナリデータ（例: 3.6MB） |
| Base64変換後 | 約 3.6MB × 4 / 3 = **4.8MB** |

さらに、HTTPリクエストヘッダ・JSONエンコード・改行などが加わるため、  
実際の送信データサイズは **5〜9MB** 程度になることがあります。

---

### 2. Finderの「MB」とコード上の「Byte数」の違い

Finderは **1MB = 1,000,000バイト（10進法）** で表示しますが、  
多くのプログラムやAPIでは **1MB = 1,048,576バイト（2進法）** を使用しています。

そのため、同じバイト数でも表記上の差が発生します。

| バイト数 | Finder表記 | プログラム表記 |
|-----------|-------------|----------------|
| 3,600,000 bytes | 約 3.6 MB | 約 3.43 MiB |

---

### 3. JSON文字列化・HTTP送信時のオーバーヘッド

APIに送信する際、以下のような要素も加わります。

- `multipart/form-data` の境界線（boundary）
- JSONオブジェクトのキー名・改行
- HTTPヘッダ情報

これらが合計して数百KB〜数MB程度増加する場合があります。

---

## 対応方法

1. **Base64変換前にサイズを確認する**
   ```python
   import os
   print(os.path.getsize("image.png"))  # バイト単位
````

2. **送信直前のBase64サイズを確認する**

   ```python
   import base64
   with open("image.png", "rb") as f:
       data = f.read()
   encoded = base64.b64encode(data)
   print(len(encoded) / 1024 / 1024, "MB")
   ```

3. **圧縮処理を行う**

   * PillowなどでJPEG圧縮（例: `quality=70`）
   * PNG→WebP変換
   * 解像度を下げる（例: `resize((1024, 1024))`）

4. **API上限を確認**

   * Dify: 約5MB〜10MBが上限（モデル・設定による）
   * Gemini: 約20MBまたは長辺4096px程度が目安（モデルごとに異なる）

---

## 参考リンク

* [Base64 Encoding – Wikipedia](https://en.wikipedia.org/wiki/Base64)
* [Dify Docs – File Upload Limitations](https://docs.dify.ai/)
* [Gemini API – Uploads and File Limits (Google AI Studio)](https://ai.google.dev/gemini-api/docs/)

---

## 結論

Finder表示の「3.6MB」と、API送信時に観測される「9.0MB」の差は、
主に **Base64エンコードによる33%増加** と **HTTP/JSON送信時のオーバーヘッド** が原因です。
送信前に画像圧縮や形式変換を行うことで、安定したAPIリクエストが可能になります。

#### 実装
```go
// 画像を圧縮する関数
func compressImage(inputPath string) (string, error) {
	log.Printf("  📥 [Compress] 画像圧縮開始: %s", inputPath)

	// 環境変数から設定を読み込み（デフォルト値あり）
	maxWidth := 1500
	quality := 85
	enableCompression := true

	if width := os.Getenv("IMAGE_MAX_WIDTH"); width != "" {
		fmt.Sscanf(width, "%d", &maxWidth)
	}
	if qual := os.Getenv("IMAGE_QUALITY"); qual != "" {
		fmt.Sscanf(qual, "%d", &quality)
	}
	if enable := os.Getenv("ENABLE_COMPRESSION"); enable == "false" {
		enableCompression = false
	}

	// 圧縮が無効の場合は元のファイルをそのまま返す
	if !enableCompression {
		log.Printf("  ⏭️  [Compress] 圧縮スキップ（ENABLE_COMPRESSION=false）")
		return inputPath, nil
	}

	// 元のファイルサイズを取得
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		log.Printf("  ❌ [Compress] ファイル情報取得失敗: %v", err)
		return "", fmt.Errorf("ファイル情報取得エラー: %v", err)
	}
	originalSize := fileInfo.Size()
	log.Printf("  🔍 [Compress] 元のサイズ: %.2f MB", float64(originalSize)/(1024*1024))

	// 画像を読み込む
	img, err := imaging.Open(inputPath)
	if err != nil {
		log.Printf("  ❌ [Compress] 画像読み込み失敗: %v", err)
		return "", fmt.Errorf("画像読み込みエラー: %v", err)
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	log.Printf("  ✅ [Compress] 画像読み込み成功: %dx%d", width, height)

	// リサイズが必要か判定
	needsResize := width > maxWidth
	var resizedImg = img

	if needsResize {
		// アスペクト比を維持してリサイズ
		newHeight := height * maxWidth / width
		resizedImg = imaging.Resize(img, maxWidth, newHeight, imaging.Lanczos)
		log.Printf("  ⚙️  [Compress] リサイズ: %dx%d -> %dx%d", width, height, maxWidth, newHeight)
	} else {
		log.Printf("  ℹ️  [Compress] リサイズ不要（幅: %dpx <= %dpx）", width, maxWidth)
	}

	// 出力ファイル名を生成
	ext := filepath.Ext(inputPath)
	baseName := strings.TrimSuffix(inputPath, ext)
	outputPath := baseName + "_compressed.jpg"

	// JPEGとして保存（品質指定）
	log.Printf("  💾 [Compress] 保存中: %s (品質: %d%%)", outputPath, quality)
	err = imaging.Save(resizedImg, outputPath, imaging.JPEGQuality(quality))
	if err != nil {
		log.Printf("  ❌ [Compress] 保存失敗: %v", err)
		return "", fmt.Errorf("画像保存エラー: %v", err)
	}

	// 圧縮後のファイルサイズを取得
	compressedInfo, err := os.Stat(outputPath)
	if err != nil {
		log.Printf("  ⚠️  [Compress] 圧縮後のファイル情報取得失敗: %v", err)
	} else {
		compressedSize := compressedInfo.Size()
		compressionRatio := float64(originalSize-compressedSize) / float64(originalSize) * 100
		log.Printf("  ✅ [Compress] 圧縮完了: %s", outputPath)
		log.Printf("  📊 [Compress] 圧縮後サイズ: %.2f MB (圧縮率: %.1f%%)",
			float64(compressedSize)/(1024*1024), compressionRatio)
	}

	return outputPath, nil
}
```