package main

import (
	"testing"
)

// TestTruncateString - 文字列切り詰めのテスト
func TestTruncateString(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"通常ケース", "これは非常に長い文字列です", 30, "これは非常に長い文字列です"},
		{"長い文字列", "これは非常に長い文字列で切り詰められます", 17, "これは非常に長い文字列で切り詰めら...(省略)"},
		{"短い文字列", "短い", 10, "短い"},
		{"空文字列", "", 10, ""},
		{"ゼロ長", "テスト", 0, "...(省略)"},
		{"英数字", "This is a very long text", 10, "This is a ...(省略)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TruncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("TruncateString() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFormatAmountWithComma - 金額フォーマットのテスト
func TestFormatAmountWithComma(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"4桁", "食費：1234", "食費：1,234"},
		{"5桁", "食費：31828", "食費：31,828"},
		{"7桁", "年収：1234567", "年収：1,234,567"},
		{"3桁以下", "食費：123", "食費：123"},
		{"不正形式", "食費1234", "食費1234"},
		{"空金額", "食費：", "食費："},
		{"文字混入", "食費：12a34", "食費：12a34"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatAmountWithComma(tt.input)
			if got != tt.want {
				t.Errorf("FormatAmountWithComma() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetMimeType - MIME type判定のテスト
func TestGetMimeType(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"JPEG", "photo.jpg", "image/jpeg"},
		{"PNG", "image.png", "image/png"},
		{"GIF", "anim.gif", "image/gif"},
		{"WebP", "modern.webp", "image/webp"},
		{"PDF", "doc.pdf", "application/pdf"},
		{"大文字", "IMAGE.JPG", "image/jpeg"},
		{"未知", "file.xyz", "application/octet-stream"},
		{"拡張子なし", "noext", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMimeType(tt.filename)
			if got != tt.want {
				t.Errorf("GetMimeType() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetPayerFromDiscordUser - Payer判定ロジックのテスト
func TestGetPayerFromDiscordUser(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		username string
		want     string
	}{
		{"ユーザーA（ID優先）", "123456789012345678", "anyname", "S"},
		{"ユーザーB（ID優先）", "796223697559748648", "anyname", "Y"},
		{"hoshi（名前）", "unknown-id", "hoshi", "S"},
		{"hoshi7hoshi（名前）", "unknown-id", "hoshi7hoshi", "Y"},
		{"未登録ユーザー", "new-id", "newuser", "S"},
		{"空", "", "", "S"},
		{"IDが優先される", "796223697559748648", "hoshi", "Y"}, // 名前ではSだがIDでY
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPayerFromDiscordUser(tt.userID, tt.username)
			if got != tt.want {
				t.Errorf("getPayerFromDiscordUser() = %v, want %v", got, tt.want)
			}
		})
	}
}
