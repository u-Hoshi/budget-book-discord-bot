# Dify × Gemini における画像サイズ制限とエラー対応まとめ

```
req_id: 1c6607052d PluginInvokeError: {"args":null,"error_type":"PluginDaemonInnerError","message":"encountered an error: invalid character '\u003c' looking for beginning of value status: 502 Bad Gateway original response: \u003chtml\u003e"}
```
上記のエラーが発生

## 🧩 概要
Dify ワークフロー内で画像をアップロードして Gemini モデルへ送信する際、  
画像サイズが大きい場合にエラーが発生する事象を確認。

本ドキュメントでは、問題の再現状況・原因・対応策・参考資料をまとめる。

---

## 🚨 問題の内容
- 約 **9.4MB の画像ファイル** を Dify ワークフロー経由で Gemini モデルに送信したところ、  
  「画像の送信に失敗」または「モデル応答なし」となる現象を確認。
- 画像を **圧縮（5MB 程度）** にしたところ、同じフローで **正常に動作** するようになった。

---

## 🔍 原因の推定

### Dify 側の制限
- Dify にはアップロードファイルサイズの上限が存在する。

| 項目 | デフォルト設定値 | 備考 |
|------|------------------|------|
| 一般ファイル | 15 MB | 複数ファイル同時アップロードも可 |
| 画像ファイル (`UPLOAD_IMAGE_FILE_SIZE_LIMIT`) | **10 MB** | セルフホスト環境変数にて設定可 |

📄 参照:  
- [Dify Docs - File Upload](https://docs.dify.ai/en/guides/workflow/file-upload)  
- [Dify Docs - Environments](https://docs.dify.ai/en/getting-started/install-self-hosted/environments)  
- [GitHub Issue #5031 - File size limit](https://github.com/langgenius/dify/issues/5031)  

---

### Gemini 側の制限
Google Gemini モデル（特に Gemini 2.0 / 2.5 Flash など）にも画像サイズ制約がある。

| 制限項目 | 上限値 | 備考 |
|-----------|--------|------|
| 1画像あたりの最大サイズ | **7 MB** | 公式ドキュメント明記 |
| 1リクエスト全体の合計サイズ | 20 MB | テキスト + 画像を含む |

📄 参照:  
- [Image understanding | Vertex AI Docs](https://cloud.google.com/vertex-ai/generative-ai/docs/multimodal/image-understanding)  
- [Gemini 2.5 Flash Model Reference](https://cloud.google.com/vertex-ai/generative-ai/docs/models/gemini/2-5-flash)  

---

## ✅ 対応方法

| 対応内容 | 詳細 |
|-----------|------|
| **画像圧縮** | アップロード前に 7MB 以下に圧縮（推奨は 5MB 前後） |
| **形式の最適化** | JPEG / WebP を使用し、不要なメタデータを除去 |
| **転送方法の見直し** | 可能な場合は「URL 参照」方式を使用し、base64 埋め込みを避ける |
| **環境変数設定 (セルフホスト)** | `UPLOAD_IMAGE_FILE_SIZE_LIMIT` を必要に応じて拡張（例: `20M`） |

---

## 🧠 補足情報
- 一部の環境（クラウド版 Dify など）では、設定値に関わらず 10MB 未満でもエラーとなることがある。  
  → ネットワーク転送時の `Content-Length` 制限や、base64 化で実際の転送量が増えるため。
- base64 エンコード後のサイズは実ファイルより約 **1.37倍** になるため、  
  **実質 7MB 制限でも、5MB 程度が安全圏**。

---

## 📚 関連ドキュメントまとめ
| 内容 | URL |
|------|-----|
| Dify File Upload 仕様 | https://docs.dify.ai/en/guides/workflow/file-upload |
| Dify 環境変数設定 | https://docs.dify.ai/en/getting-started/install-self-hosted/environments |
| Dify Issue - ファイルサイズ制限 | https://github.com/langgenius/dify/issues/5031 |
| Gemini Image Understanding 仕様 | https://cloud.google.com/vertex-ai/generative-ai/docs/multimodal/image-understanding |
| Gemini 2.5 Flash モデル仕様 | https://cloud.google.com/vertex-ai/generative-ai/docs/models/gemini/2-5-flash |

---

## 💡 今後の推奨運用
- 画像送信前に圧縮を自動化する関数／ユーティリティを導入する。
- Dify のワークフロー入力時に、ファイルサイズをチェックし警告を出す処理を追加する。
- 定期的に Gemini / Dify の最新仕様を確認し、上限変更に備える。

---

### 🧾 更新履歴
| 日付 | 内容 |
|------|------|
| 2025-11-02 | 初版作成（原因分析・対応策まとめ） |
