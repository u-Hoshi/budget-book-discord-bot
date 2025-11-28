package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// 添付画像をローカルに保存する関数
func DownloadImage(url, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("❌ HTTPリクエスト失敗: %v", err)
		return err
	}
	defer resp.Body.Close()

	// 一時ディレクトリを使用
	tempFile := filepath.Join(os.TempDir(), filename)
	out, err := os.Create(tempFile)
	if err != nil {
		log.Printf("❌ 一時ファイル作成失敗: %v", err)
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Printf("❌ ファイル書き込み失敗: %v", err)
		return err
	}

	return nil
}

// 画像を圧縮する関数
func CompressImage(inputPath string) (string, error) {
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
		return inputPath, nil
	}

	// 一時ディレクトリ内のファイルパスに変更
	tempInputPath := filepath.Join(os.TempDir(), filepath.Base(inputPath))

	// 元のファイルサイズを取得
	fileInfo, err := os.Stat(tempInputPath)
	if err != nil {
		log.Printf("❌ ファイル情報取得失敗: %v", err)
		return "", fmt.Errorf("ファイル情報取得エラー: %v", err)
	}
	originalSize := fileInfo.Size()

	// 画像を読み込む
	img, err := imaging.Open(tempInputPath)
	if err != nil {
		log.Printf("❌ 画像読み込み失敗: %v", err)
		return "", fmt.Errorf("画像読み込みエラー: %v", err)
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// リサイズが必要か判定
	needsResize := width > maxWidth
	var resizedImg = img

	if needsResize {
		// アスペクト比を維持してリサイズ
		newHeight := height * maxWidth / width
		resizedImg = imaging.Resize(img, maxWidth, newHeight, imaging.Lanczos)
	}

	// 一時ディレクトリに出力ファイル名を生成
	ext := filepath.Ext(tempInputPath)
	baseName := strings.TrimSuffix(filepath.Base(tempInputPath), ext)
	outputPath := filepath.Join(os.TempDir(), baseName+"_compressed.jpg")

	// JPEGとして保存（品質指定）
	err = imaging.Save(resizedImg, outputPath, imaging.JPEGQuality(quality))
	if err != nil {
		log.Printf("❌ 画像保存失敗: %v", err)
		return "", fmt.Errorf("画像保存エラー: %v", err)
	}

	// 圧縮後のファイルサイズを取得
	compressedInfo, err := os.Stat(outputPath)
	if err == nil {
		compressedSize := compressedInfo.Size()
		compressionRatio := float64(originalSize-compressedSize) / float64(originalSize) * 100
		log.Printf("✅ 画像圧縮完了: %.1f%% 削減", compressionRatio)
	}

	return outputPath, nil
}
