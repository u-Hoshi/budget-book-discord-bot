package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

// DifyのレスポンスJSON構造体
type DifyFileUploadResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Extension string `json:"extension"`
	MimeType  string `json:"mime_type"`
	CreatedBy string `json:"created_by"`
	CreatedAt int64  `json:"created_at"`
}

type DifyWorkflowResponse struct {
	WorkflowRunID string                 `json:"workflow_run_id"`
	TaskID        string                 `json:"task_id"`
	Data          map[string]interface{} `json:"data"`
}

// 画像をDifyにアップロードする関数
func UploadImageToDify(filename string) (string, error) {
	log.Printf("Difyへのアップロード開始: %s", filename)

	difyToken := os.Getenv("DIFY_API_KEY")
	// DIFY_ENDPOINTとDIFY_API_URLの両方をサポート（後方互換性）
	difyEndpoint := os.Getenv("DIFY_ENDPOINT")
	if difyEndpoint == "" {
		difyEndpoint = os.Getenv("DIFY_API_URL")
	}

	if difyToken == "" {
		log.Printf("❌ DIFY_API_KEYが未設定")
		return "", fmt.Errorf("DIFY_API_KEYが設定されていません")
	}

	// 空白をトリミング
	difyToken = strings.TrimSpace(difyToken)

	if difyEndpoint == "" {
		difyEndpoint = "https://api.dify.ai/v1" // デフォルト値
		log.Printf("エンドポイント未設定、デフォルト使用: %s", difyEndpoint)
	}

	// ファイルを開く
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("❌ ファイルオープン失敗: %v", err)
		return "", fmt.Errorf("ファイルオープンエラー: %v", err)
	}
	defer file.Close()

	// multipart/form-dataを作成
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// ファイル拡張子からMIME typeを判定
	mimeType := GetMimeType(filename)

	// Content-Dispositionヘッダーを手動で作成
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filepath.Base(filename)))
	h.Set("Content-Type", mimeType)

	part, err := writer.CreatePart(h)
	if err != nil {
		log.Printf("❌ フォームパート作成失敗: %v", err)
		return "", fmt.Errorf("フォームパート作成エラー: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		log.Printf("❌ ファイルコピー失敗: %v", err)
		return "", fmt.Errorf("ファイルコピーエラー: %v", err)
	}

	// userフィールドを追加
	_ = writer.WriteField("user", "discord-bot-user")

	err = writer.Close()
	if err != nil {
		log.Printf("❌ writer close失敗: %v", err)
		return "", fmt.Errorf("writer closeエラー: %v", err)
	}

	// リクエストを作成
	uploadURL := fmt.Sprintf("%s/files/upload", difyEndpoint)
	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		log.Printf("❌ リクエスト作成失敗: %v", err)
		return "", fmt.Errorf("リクエスト作成エラー: %v", err)
	}

	// ヘッダーを設定
	contentType := writer.FormDataContentType()
	authHeader := "Bearer " + difyToken

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", authHeader)

	// リクエストを送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ リクエスト送信失敗: %v", err)
		return "", fmt.Errorf("リクエスト送信エラー: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスを読み取る
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ レスポンス読み取り失敗: %v", err)
		return "", fmt.Errorf("レスポンス読み取りエラー: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("❌ アップロード失敗 - ステータス: %d", resp.StatusCode)

		// 401エラーの場合は認証問題を指摘
		if resp.StatusCode == 401 {
			log.Printf("認証エラー: API Keyの設定を確認してください")
		}

		return "", fmt.Errorf("アップロード失敗 (ステータス: %d): %s", resp.StatusCode, string(respBody))
	}

	// JSONをパース
	var uploadResp DifyFileUploadResponse
	err = json.Unmarshal(respBody, &uploadResp)
	if err != nil {
		log.Printf("❌ JSONパース失敗: %v", err)
		return "", fmt.Errorf("JSONパースエラー: %v, レスポンス: %s", err, string(respBody))
	}

	// log.Printf("✅ アップロード成功 - ID: %s", uploadResp.ID)
	return uploadResp.ID, nil
}

// DifyのワークフローまたはチャットBotに画像を送信して処理を実行する関数
func RunDifyWorkflowWithImage(fileID, userID, username string) (string, error) {
	log.Printf("🚀 Difyワークフロー実行開始 - UserID: %s, Username: %s, FileID: %s", userID, username, fileID)

	difyToken := os.Getenv("DIFY_API_KEY")
	// DIFY_ENDPOINTとDIFY_API_URLの両方をサポート（後方互換性）
	difyEndpoint := os.Getenv("DIFY_ENDPOINT")
	if difyEndpoint == "" {
		difyEndpoint = os.Getenv("DIFY_API_URL")
	}
	difyWorkflowID := os.Getenv("DIFY_WORKFLOW_ID") // ワークフローの場合
	difyInputName := os.Getenv("DIFY_INPUT_NAME")   // input変数名（デフォルト: receipt_images）
	if difyInputName == "" {
		difyInputName = "receipt_images" // デフォルト値
	}

	if difyToken == "" {
		log.Printf("❌ DIFY_API_KEYが未設定")
		return "", fmt.Errorf("DIFY_API_KEYが設定されていません")
	}

	// 空白をトリミング
	difyToken = strings.TrimSpace(difyToken)

	if difyEndpoint == "" {
		difyEndpoint = "https://api.dify.ai/v1"
	}

	// ワークフローを実行する場合
	// inputs に画像のfile_idを含める
	// Difyワークフローが期待する形式で画像データを作成
	imageData := map[string]interface{}{
		"transfer_method": "local_file",
		"upload_file_id":  fileID,
		"type":            "image",
	}

	// DiscordユーザーからPayerを判定
	payer := getPayerFromDiscordUser(userID, username)
	log.Printf("🔑 判定されたPayer: %s (UserID: %s, Username: %s)", payer, userID, username)

	requestBody := map[string]interface{}{
		"inputs": map[string]interface{}{
			difyInputName: []interface{}{imageData}, // 配列形式で送信
			"payer":       payer,                    // "Y" または "S" を直接送信
		},
		"response_mode": "blocking", // または "streaming"
		"user":          "discord-bot-user",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("❌ JSONマーシャル失敗: %v", err)
		return "", fmt.Errorf("JSONマーシャルエラー: %v", err)
	}

	// デバッグ用: 送信するJSONをログ出力
	log.Printf("📤 Difyへ送信するJSON: %s", string(jsonData))

	// APIエンドポイントを決定
	var apiURL string
	if difyWorkflowID != "" {
		// ワークフローを使用する場合
		apiURL = fmt.Sprintf("%s/workflows/run", difyEndpoint)
	} else {
		apiURL = fmt.Sprintf("%s/workflows/run", difyEndpoint)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("❌ リクエスト作成失敗: %v", err)
		return "", fmt.Errorf("リクエスト作成エラー: %v", err)
	}

	authHeader := "Bearer " + difyToken

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)

	// リクエストを送信
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ リクエスト送信失敗: %v", err)
		return "", fmt.Errorf("リクエスト送信エラー: %v", err)
	}
	defer resp.Body.Close()

	// レスポンスを読み取る
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ レスポンス読み取り失敗: %v", err)
		return "", fmt.Errorf("レスポンス読み取りエラー: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		log.Printf("❌ ワークフロー実行失敗 - ステータス: %d, UserID: %s, Payer: %s", resp.StatusCode, userID, getPayerFromDiscordUser(userID, username))
		log.Printf("📥 Difyからのエラーレスポンス: %s", string(respBody))

		// 400エラーの場合は入力パラメータの問題を指摘
		if resp.StatusCode == 400 {
			log.Printf("リクエストパラメータエラー: Difyワークフローの設定を確認してください")
		}

		// 500エラーの場合はDifyサーバー側の問題を指摘
		if resp.StatusCode == 500 {
			log.Printf("⚠️  Difyサーバー内部エラー: ワークフロー内のロジックやプラグインを確認してください")
		}

		return "", fmt.Errorf("ワークフロー実行失敗 (ステータス: %d): %s", resp.StatusCode, string(respBody))
	}

	log.Printf("✅ ワークフロー実行成功")

	// レスポンスをパースしてエラーをチェック
	var workflowResp map[string]interface{}
	err = json.Unmarshal(respBody, &workflowResp)
	if err != nil {
		log.Printf("⚠️  レスポンスのJSONパースに失敗: %v", err)
		return string(respBody), nil // パースできなくてもレスポンスは返す
	}

	// Dify内部エラーをチェック
	if errorData, hasError := workflowResp["error"]; hasError {
		log.Printf("⚠️  Dify内部エラーを検出: %v", errorData)

		// PluginDaemonInnerErrorの場合
		if strings.Contains(fmt.Sprintf("%v", errorData), "PluginDaemonInnerError") {
			log.Printf("Difyワークフロー内のプラグインでエラーが発生しました。管理画面でワークフローのログを確認してください。")
		}
	}

	return string(respBody), nil
}
