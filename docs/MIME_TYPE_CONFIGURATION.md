# MIME Type設定ガイド

## 問題: Difyでの画像認識エラー

### 現象
Discord経由でアップロードした画像のMIME typeが`application/octet-stream`になり、Difyが画像として認識できない。

### 原因
`multipart.Writer.CreateFormFile()`を使用すると、MIME typeが自動的に`application/octet-stream`に設定されてしまう。

## 解決方法

### 実装内容

ファイル拡張子に基づいて適切なMIME typeを自動判定し、明示的に設定するように変更しました。

```go
// ファイル拡張子からMIME typeを判定
mimeType := getMimeType(filename)

// Content-Dispositionヘッダーを手動で作成
h := make(textproto.MIMEHeader)
h.Set("Content-Disposition", `form-data; name="file"; filename="..."`)
h.Set("Content-Type", mimeType)

part, err := writer.CreatePart(h)
```

### サポートされるMIME type

#### 画像形式
| 拡張子 | MIME type |
|--------|-----------|
| .jpg, .jpeg | image/jpeg |
| .png | image/png |
| .gif | image/gif |
| .bmp | image/bmp |
| .webp | image/webp |
| .svg | image/svg+xml |
| .ico | image/x-icon |

#### その他
| 拡張子 | MIME type |
|--------|-----------|
| .pdf | application/pdf |
| .txt | text/plain |
| .csv | text/csv |
| .json | application/json |
| .xml | application/xml |

未対応の拡張子の場合は `application/octet-stream` がデフォルトで使用されます。

## ログ出力

実装後は以下のようなログが出力されます：

```log
📝 [Dify Upload] フォームフィールド追加中: file=IMG_0388.JPG
🔍 [Dify Upload] 検出されたMIME type: image/jpeg
✅ [Dify Upload] ファイルコピー完了: 123456 bytes (MIME type: image/jpeg)
```

## 確認方法

### 1. Dify側でのMIME type確認

Difyのファイルアップロードレスポンスに含まれる`mime_type`を確認：

```json
{
  "id": "abc123",
  "name": "IMG_0388.JPG",
  "size": 123456,
  "mime_type": "image/jpeg"  ← ここを確認
}
```

### 2. Botのログで確認

```log
✅ [Dify Upload] アップロード成功 - ID: abc123, Name: IMG_0388.JPG, Size: 123456
📄 [Dify Upload] レスポンスボディ: {"mime_type":"image/jpeg",...}
```

### 3. Dify管理画面で確認

1. Difyの管理画面にログイン
2. **Files**または**Assets**セクションを確認
3. アップロードされたファイルのMIME typeを確認

## トラブルシューティング

### 問題1: まだ`application/octet-stream`になる

**原因**:
- ファイル拡張子が認識されていない
- 拡張子が大文字小文字で異なる

**確認**:
```log
🔍 [Dify Upload] 検出されたMIME type: application/octet-stream
```

**解決方法**:
1. ファイル名を確認（例: `IMG_0388.JPG`）
2. 拡張子が`.jpg`か`.jpeg`か確認
3. 必要に応じて`getMimeType()`関数に拡張子を追加

### 問題2: 特定の画像形式が認識されない

**例**: `.heic`（iPhoneのデフォルト形式）

**解決方法**:
`main.go`の`getMimeType()`関数に追加：

```go
".heic": "image/heic",
".heif": "image/heif",
```

### 問題3: Difyで画像として認識されない

**確認項目**:
1. MIME typeが正しく設定されているか
2. ファイルが破損していないか
3. Difyワークフローの入力設定が正しいか
   - Type: `File`
   - File Type: `Image`

## 新しい形式の追加方法

新しいファイル形式をサポートする場合は、`getMimeType()`関数に追加してください。

### 例: HEIC形式を追加

```go
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	mimeTypes := map[string]string{
		// 既存の設定...
		
		// 追加
		".heic": "image/heic",
		".heif": "image/heif",
		".avif": "image/avif",
	}
	
	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}
	
	return "application/octet-stream"
}
```

## 動作確認

### テスト手順

1. **異なる画像形式でテスト**
   ```
   - test.jpg  → image/jpeg
   - test.png  → image/png
   - test.gif  → image/gif
   ```

2. **Botを起動**
   ```bash
   go run main.go
   ```

3. **Discord経由でアップロード**
   - 各形式の画像を添付
   - `!upload`コマンドを実行

4. **ログを確認**
   ```log
   🔍 [Dify Upload] 検出されたMIME type: image/jpeg
   ✅ [Dify Upload] ファイルコピー完了: 123456 bytes (MIME type: image/jpeg)
   ```

5. **Difyレスポンスを確認**
   ```json
   {
     "mime_type": "image/jpeg"
   }
   ```

## 期待される結果

### 修正前
```json
{
  "id": "abc123",
  "mime_type": "application/octet-stream",  ❌
  "name": "IMG_0388.JPG"
}
```

### 修正後
```json
{
  "id": "abc123",
  "mime_type": "image/jpeg",  ✅
  "name": "IMG_0388.JPG"
}
```

## 参考情報

### MIME type一覧
- [IANA Media Types](https://www.iana.org/assignments/media-types/media-types.xhtml)
- [MDN - MIME types](https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types)

### RFC
- [RFC 2046 - Multipurpose Internet Mail Extensions (MIME)](https://www.rfc-editor.org/rfc/rfc2046.html)

## 更新履歴

- 2025/11/02: 初版作成 - MIME type自動判定機能を実装
