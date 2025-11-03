# Difyワークフロー input変数の設定方法

## 問題: 400 Bad Request エラー

### エラーメッセージ
```json
{
  "code": "invalid_param",
  "message": "receipt_images is required in input form",
  "status": 400
}
```

このエラーは、Difyワークフローが期待しているinput変数名と、Botが送信している変数名が一致していない場合に発生します。

## 解決方法

### 1. Difyワークフローのinput変数名を確認

1. [Dify管理画面](https://cloud.dify.ai/)にログイン
2. 対象のワークフローを開く
3. **Start** ノードをクリック
4. **Variables**（変数）セクションを確認
5. 画像入力用の変数名をメモ（例: `receipt_images`, `image`, `images`など）

### 2. .envファイルに変数名を設定

`.env`ファイルに以下を追加：

```env
DIFY_INPUT_NAME=receipt_images
```

**変数名の例**:
- `receipt_images` - レシート画像
- `image` - 単一画像
- `images` - 複数画像
- `file` - ファイル
- その他、ワークフローで定義した名前

### 3. ワークフロー側の設定確認

Difyワークフローの**Start**ノードで以下を確認：

#### 変数の設定
- **Variable Name**: `receipt_images`（または任意の名前）
- **Type**: `File`
- **File Type**: `Image`
- **Is Array**: ✅（チェックを入れる）← **重要！**

#### 設定例
```
Variable Name: receipt_images
Type: File
File Type: Image
Is Array: Yes (複数ファイルを受け取る場合)
Required: Yes
```

## Botの実装

現在のBotは以下の形式でデータを送信しています：

```json
{
  "inputs": {
    "receipt_images": [
      {
        "transfer_method": "local_file",
        "upload_file_id": "ファイルID",
        "type": "image"
      }
    ]
  },
  "response_mode": "blocking",
  "user": "discord-bot-user"
}
```

### データ形式の説明

1. **配列形式**: `[]` で囲まれている
   - Difyは画像を配列として受け取る
   - 単一の画像でも配列形式で送信

2. **transfer_method**: `local_file`
   - アップロード済みのファイルを参照

3. **upload_file_id**: アップロードAPIで取得したID

4. **type**: `image`
   - ファイルタイプを明示

## トラブルシューティング

### ケース1: 変数名が間違っている

**エラー**:
```
"message": "receipt_images is required in input form"
```

**解決方法**:
1. Difyワークフローの変数名を確認
2. `.env`の`DIFY_INPUT_NAME`を修正
3. Botを再起動

### ケース2: 配列形式になっていない

**エラー**:
```
"message": "Expected array but got object"
```

**解決方法**:
- 現在のBotは自動的に配列形式で送信します
- ワークフロー側で`Is Array: Yes`に設定されているか確認

### ケース3: ファイルタイプが合わない

**エラー**:
```
"message": "Invalid file type"
```

**解決方法**:
- ワークフローの`File Type`が`Image`に設定されているか確認
- アップロードしたファイルが画像形式（jpg, png, gif等）か確認

## デバッグログの確認

Bot実行時に以下のログが出力されます：

```log
📝 [Dify Workflow] Input変数名: receipt_images
📋 [Dify Workflow] inputs構造:
  - receipt_images: 配列（要素数: 1）
  - transfer_method: local_file
  - upload_file_id: abc123...
```

400エラー時:
```log
🔍 [Dify Workflow] 400エラー診断:
  ❌ リクエストパラメータが不正です。以下を確認してください:
  1. Difyワークフローのinput変数名が 'receipt_images' か確認
  2. ワークフローのinput typeが 'File' (配列) に設定されているか
  3. FileIDが正しく渡されているか
  4. 送信したJSON: {...}
```

## 環境変数一覧

```env
# 必須
DIFY_API_KEY=app-xxxxxxxxxxxx
DIFY_ENDPOINT=https://api.dify.ai/v1

# オプション（デフォルト値あり）
DIFY_INPUT_NAME=receipt_images  # デフォルト: receipt_images
DIFY_WORKFLOW_ID=               # ワークフローID（将来的に使用予定）
```

## 正常な動作ログ

```log
✅ [STEP 6/8 完了] Difyアップロード成功 - FileID: abc123
🔄 [STEP 7/8] Difyワークフロー実行開始 - FileID: abc123
  📝 [Dify Workflow] Input変数名: receipt_images
  📋 [Dify Workflow] inputs構造:
    - receipt_images: 配列（要素数: 1）
    - transfer_method: local_file
    - upload_file_id: abc123
  📤 [Dify Workflow] HTTPリクエスト送信中...
  📥 [Dify Workflow] レスポンス受信 - Status: 200 OK (200)
  ✅ [Dify Workflow] ワークフロー実行成功
✅ [STEP 7/8 完了] Difyワークフロー実行成功
```

## 参考リンク

- [Dify Workflow Documentation](https://docs.dify.ai/guides/workflow)
- [Dify API Reference](https://docs.dify.ai/guides/application-publishing/developing-with-apis)
