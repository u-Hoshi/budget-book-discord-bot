package main

import (
	"path/filepath"
	"strings"
)

func TruncateString(s string, maxLen int) string {
	// ルーン（文字）で長さを判定（マルチバイト文字対応）
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "...(省略)"
}

// 金額にカンマを追加する関数（例: "食費：31828" -> "食費：31,828"）
func FormatAmountWithComma(s string) string {
	// "項目：金額"の形式を分割
	parts := strings.Split(s, "：")
	if len(parts) != 2 {
		return s // 形式が異なる場合はそのまま返す
	}

	category := parts[0]
	amountStr := strings.TrimSpace(parts[1])

	// 金額が空の場合はそのまま返す
	if amountStr == "" {
		return s
	}

	// 数値以外が含まれている場合はそのまま返す
	for _, c := range amountStr {
		if c < '0' || c > '9' {
			return s
		}
	}

	// 3桁ごとにカンマを挿入
	var result strings.Builder
	n := len(amountStr)
	for i, digit := range amountStr {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}

	return category + "：" + result.String()
}

// ファイル名からMIME typeを判定する
func GetMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	// 画像形式
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".ico":  "image/x-icon",
		// PDFなど
		".pdf": "application/pdf",
		// テキスト
		".txt":  "text/plain",
		".csv":  "text/csv",
		".json": "application/json",
		".xml":  "application/xml",
	}

	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}

	// デフォルト
	return "application/octet-stream"
}
