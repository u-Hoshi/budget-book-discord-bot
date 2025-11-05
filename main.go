package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/disintegration/imaging"
	"github.com/joho/godotenv"
)

// ヘルスチェック用のレスポンス構造体
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
	Uptime    string    `json:"uptime,omitempty"`
}

var startTime = time.Now()

// ヘルスチェックエンドポイント
func healthHandler(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(startTime)

	health := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now(),
		Version:   "1.0.0",
		Uptime:    uptime.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

// HTTPサーバーを開始する関数
func startHTTPServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", healthHandler)

	// 環境変数からポートを取得（デフォルト: 8080）
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		log.Printf("🌐 HTTPサーバーを開始: ポート %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTPサーバーエラー: %v", err)
		}
	}()

	return server
}

// 定期的なヘルスチェック機能
func startHealthCheckCron() {
	// 環境変数からヘルスチェックURLを取得
	healthCheckURL := os.Getenv("HEALTH_CHECK_URL")
	if healthCheckURL == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		healthCheckURL = fmt.Sprintf("http://localhost:%s", port)
	}

	log.Printf("🕐 ヘルスチェックの定期実行を開始しました (5分間隔)")
	log.Printf("🔗 ヘルスチェックURL: %s", healthCheckURL)

	// 初回ヘルスチェック（5秒後に実行）
	go func() {
		time.Sleep(5 * time.Second)
		performHealthCheck(healthCheckURL)
	}()

	// 5分間隔のティッカーを作成
	ticker := time.NewTicker(5 * time.Minute)

	go func() {
		defer ticker.Stop()

		for range ticker.C {
			performHealthCheck(healthCheckURL)
		}
	}()
}

// ヘルスチェックを実行する関数
func performHealthCheck(url string) {
	now := time.Now().Format("2006-01-02 15:04:05")
	log.Printf("🔍 [%s] ヘルスチェック実行中... (%s)", now, url)

	// タイムアウト付きのHTTPクライアント
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("❌ [%s] ヘルスチェックエラー: %v", now, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		log.Printf("✅ [%s] ヘルスチェック成功: %d", now, resp.StatusCode)
	} else {
		log.Printf("⚠️ [%s] ヘルスチェック失敗: %d", now, resp.StatusCode)
	}
} // 文字列を指定した長さに切り詰める
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(省略)"
}

// ファイル名からMIME typeを判定する
func getMimeType(filename string) string {
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

// DiscordユーザーIDまたはユーザー名からpayerを判定する関数
func getPayerFromDiscordUser(userID, username string) string {
	// ユーザーIDで判定（優先）
	switch userID {
	case "123456789012345678": // 例: ユーザーAのID
		return "S"
	case "796223697559748648": // 例: ユーザーBのID
		return "Y"
	}

	// ユーザー名で判定（フォールバック）
	switch username {
	case "hoshi":
		return "S"
	case "hoshi7hoshi":
		return "Y"
	}

	// デフォルト値
	log.Printf("未登録ユーザー（ID: %s, Username: %s） -> デフォルトPayer: S", userID, username)
	return "S"
}

// 画像を圧縮する関数
func compressImage(inputPath string) (string, error) {
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

// 既存: 単体→複数定義へ
var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "hello",
		Description: "挨拶を返します",
	},
}

func main() {
	log.Println("🚀 Discord Bot 起動中...")

	// .envファイルの読み込み
	err := godotenv.Load()
	if err != nil {
		log.Printf("⚠️  .envファイルの読み込みに失敗しました（環境変数から読み込みます）: %v", err)
	}

	// 環境変数の確認
	appID := os.Getenv(("APPLICATION_ID"))
	token := os.Getenv("DISCORD_TOKEN")
	difyAPIKey := os.Getenv("DIFY_API_KEY")

	if appID == "" {
		log.Fatal("❌ APPLICATION_IDが未設定です。")
	}
	if token == "" {
		log.Fatal("❌ DISCORD TOKENが未設定です。")
	}
	if difyAPIKey == "" {
		log.Println("⚠️  DIFY_API_KEYが未設定です。画像アップロード機能は使用できません。")
	}
	log.Println("✅ 必要な環境変数が設定されています。")
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("セッションの作成に失敗しました: %v", err)
	}

	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	// メッセージ受信時のハンドラを追加
	dg.AddHandler(onMessageCreate)

	// スラッシュコマンドのハンドラ
	dg.AddHandler((func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}
		var response *discordgo.InteractionResponse
		switch i.ApplicationCommandData().Name {

		case "hello":
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "やっほー‼️‼️‼️",
				},
			})
			_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Are you comfortable with buttons and other message components?",
					Flags:   discordgo.MessageFlagsEphemeral,
					// Buttons and other components are specified in Components field.
					Components: []discordgo.MessageComponent{
						// ActionRow is a container of all buttons within the same row.
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									// Label is what the user will see on the button.
									Label: "Yes",
									// Style provides coloring of the button. There are not so many styles tho.
									Style: discordgo.SuccessButton,
									// Disabled allows bot to disable some buttons for users.
									Disabled: false,
									// CustomID is a thing telling Discord which data to send when this button will be pressed.
									CustomID: "fd_yes",
								},
								discordgo.Button{
									Label:    "No",
									Style:    discordgo.DangerButton,
									Disabled: false,
									CustomID: "fd_no",
								},
								discordgo.Button{
									Label:    "I don't know",
									Style:    discordgo.LinkButton,
									Disabled: false,
									// Link buttons don't require CustomID and do not trigger the gateway/HTTP event
									URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
									Emoji: &discordgo.ComponentEmoji{
										Name: "🤷",
									},
								},
							},
						},
						// The message may have multiple actions rows.
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.Button{
									Label:    "Discord Developers server",
									Style:    discordgo.LinkButton,
									Disabled: false,
									URL:      "https://discord.gg/discord-developers",
								},
							},
						},
					},
				},
			})
		}

		err := s.InteractionRespond(i.Interaction, response)
		if err != nil {
			panic(err)
		}
	}))

	// グローバル登録 (複数ループ)
	for _, c := range commands {
		newCmd, err := dg.ApplicationCommandCreate(appID, "", c)
		if err != nil {
			log.Fatalf("コマンド登録失敗 (%s): %v", c.Name, err)
		}
		log.Printf("登録: /%s\n", newCmd.Name)
	}

	// アプリケーション起動時に呼ばれる

	if err = dg.Open(); err != nil {
		log.Fatalf("接続エラー: %v", err)
	}

	defer dg.Close()

	// HTTPサーバーを開始
	httpServer := startHTTPServer()

	// ヘルスチェック機能を開始
	startHealthCheckCron()

	log.Println("✅ Bot起動完了 - Ctrl+Cで終了")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("🔄 シャットダウン開始...")

	// HTTPサーバーを graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTPサーバーのシャットダウンエラー: %v", err)
	} else {
		log.Println("✅ HTTPサーバーを正常に停止しました")
	}

	log.Println("✅ 終了完了")
}

// メッセージを受け取った時のイベントハンドラ
func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// 自分のメッセージは無視
	if m.Author.ID == s.State.User.ID {
		return
	}

	// !pingコマンド
	if m.Content == "!ping" {
		_, _ = s.ChannelMessageSend(m.ChannelID, "Pong!")
		return
	}

	// !whoamiコマンド
	if m.Content == "!whoami" {
		userInfo := fmt.Sprintf("👤 **あなたの情報**\n```\nユーザーID: %s\nユーザー名: %s\n表示名: %s\n```\n💡 この情報を使ってPayerを設定できます！",
			m.Author.ID, m.Author.Username, m.Author.GlobalName)
		_, _ = s.ChannelMessageSend(m.ChannelID, userInfo)
		return
	}

	// "!upload" で呼び出し
	// if m.Content == "!upload" {
	// Bot自身のメッセージは無視
	if m.Author.Bot {
		return
	}

	// 対象チャンネルID
	const targetChannelID = "1435607678029140078"

	// 対象チャンネル以外は無視
	if m.ChannelID != targetChannelID {
		return
	}

	// 添付ファイルがある（＝画像などが投稿された）
	if len(m.Attachments) > 0 {
		log.Printf("📷 画像アップロード処理開始 - User: %s, 画像数: %d", m.Author.Username, len(m.Attachments))

		// 処理開始メッセージ
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🖼️ %d個の画像を処理中です...", len(m.Attachments)))

		// 全ての添付ファイルを処理
		successCount := 0
		failureCount := 0

		for i, attachment := range m.Attachments {
			log.Printf("📎 [%d/%d] 処理中: %s", i+1, len(m.Attachments), attachment.Filename)

			imageURL := attachment.URL
			fileName := attachment.Filename

			// 各画像の処理状況をログ出力
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("� [%d/%d] %s を処理中...", i+1, len(m.Attachments), fileName))

			// 一時保存する場合（例: difyなどにPOST前にローカルで保持したい）
			err := downloadImage(imageURL, fileName)
			if err != nil {
				log.Printf("❌ [%d/%d] 画像ダウンロード失敗 (%s): %v", i+1, len(m.Attachments), fileName, err)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ [%d/%d] %s のダウンロードに失敗しました: %v", i+1, len(m.Attachments), fileName, err))
				failureCount++
				continue
			}

			// 一時ディレクトリ内のファイルパスを取得
			tempFilePath := filepath.Join(os.TempDir(), fileName)

			// --- 画像を圧縮 ---
			compressedFileName, err := compressImage(tempFilePath)
			if err != nil {
				log.Printf("❌ [%d/%d] 画像圧縮失敗 (%s): %v", i+1, len(m.Attachments), fileName, err)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ [%d/%d] %s の圧縮に失敗しました: %v", i+1, len(m.Attachments), fileName, err))
				os.Remove(tempFilePath)
				failureCount++
				continue
			}

			// 元の画像ファイルを削除（圧縮版を使用）
			if compressedFileName != tempFilePath {
				os.Remove(tempFilePath)
			}

			// --- Dify APIに送信 ---
			// 1. 画像をDifyにアップロード
			fileID, err := uploadImageToDify(compressedFileName)
			if err != nil {
				log.Printf("❌ [%d/%d] Difyアップロード失敗 (%s): %v", i+1, len(m.Attachments), fileName, err)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ [%d/%d] %s のDifyアップロードに失敗しました: %v", i+1, len(m.Attachments), fileName, err))
				os.Remove(compressedFileName)
				failureCount++
				continue
			}

			// 2. ワークフローを実行（画像を使用）
			result, err := runDifyWorkflowWithImage(fileID, m.Author.ID, m.Author.Username)
			if err != nil {
				log.Printf("❌ [%d/%d] Difyワークフロー実行失敗 (%s): %v", i+1, len(m.Attachments), fileName, err)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ [%d/%d] %s のDify処理に失敗しました: %v", i+1, len(m.Attachments), fileName, err))
				os.Remove(compressedFileName)
				failureCount++
				continue
			}

			// 成功メッセージ
			// レスポンスをパースして結果を整形
			var resultData map[string]interface{}
			if err := json.Unmarshal([]byte(result), &resultData); err == nil {
				// エラーがあるかチェック
				if errorMsg, hasError := resultData["error"]; hasError {
					errorStr := fmt.Sprintf("%v", errorMsg)
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ [%d/%d] %s: Difyワークフローは実行されましたが、内部でエラーが発生しました。\n```\n%s\n```", i+1, len(m.Attachments), fileName, truncateString(errorStr, 800)))
					failureCount++
				} else {
					// 正常な結果を表示
					if data, hasData := resultData["data"]; hasData {
						// outputs.output配列から店舗・金額・項目を抽出
						var store, item string
						var amount int
						var display string
						if outputs, ok := resultData["outputs"].(map[string]interface{}); ok {
							if outputArr, ok := outputs["output"].([]interface{}); ok && len(outputArr) > 0 {
								// 1つ目の要素をJSONとしてパース
								var outputObj map[string]interface{}
								// outputArr[0]はstring型のJSON
								if str, ok := outputArr[0].(string); ok {
									if err := json.Unmarshal([]byte(str), &outputObj); err == nil {
										if inserted, ok := outputObj["insertedData"].(map[string]interface{}); ok {
											if v, ok := inserted["store"].(string); ok {
												store = v
											}
											if v, ok := inserted["item"].(string); ok {
												item = v
											}
											if v, ok := inserted["amount"].(float64); ok {
												amount = int(v)
											}
											display = fmt.Sprintf("✅ [%d/%d] %s: Dify処理が完了しました！\n店舗: %s\n金額: %d円\n項目: %s", i+1, len(m.Attachments), fileName, store, amount, item)
										}
									}
								}
							}
						}
						if display != "" {
							s.ChannelMessageSend(m.ChannelID, display)
						} else {
							dataJSON, _ := json.MarshalIndent(data, "", "  ")
							s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ [%d/%d] %s: Dify処理が完了しました！\n```json\n%s\n```", i+1, len(m.Attachments), fileName, truncateString(string(dataJSON), 1200)))
						}
					} else {
						s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ [%d/%d] %s: Dify処理が完了しました！\n```json\n%s\n```", i+1, len(m.Attachments), fileName, truncateString(result, 1200)))
					}
					successCount++
				}
			} else {
				// JSONパースできない場合はそのまま表示
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ [%d/%d] %s: Dify処理が完了しました！\n```\n%s\n```", i+1, len(m.Attachments), fileName, truncateString(result, 1200)))
				successCount++
			}

			// 一時ファイルを削除
			err = os.Remove(compressedFileName)
			if err != nil {
				log.Printf("⚠️ [%d/%d] 一時ファイルの削除に失敗 (%s): %v", i+1, len(m.Attachments), fileName, err)
			}

			log.Printf("✅ [%d/%d] 画像処理が完了しました: %s", i+1, len(m.Attachments), fileName)

			// 複数画像処理時は適度に間隔を空ける（最後の画像以外）
			if i < len(m.Attachments)-1 {
				time.Sleep(2 * time.Second)
				log.Printf("⏱️ 次の画像処理まで2秒待機...")
			}
		}

		// 全体の処理結果をサマリー表示
		totalImages := len(m.Attachments)
		if successCount == totalImages {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🎉 全ての画像処理が完了しました！\n✅ 成功: %d個\n❌ 失敗: %d個", successCount, failureCount))
		} else if successCount > 0 {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ 一部の画像処理が完了しました。\n✅ 成功: %d個\n❌ 失敗: %d個", successCount, failureCount))
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ 全ての画像処理が失敗しました。\n✅ 成功: %d個\n❌ 失敗: %d個", successCount, failureCount))
		}

		log.Printf("📊 画像処理サマリー - 成功: %d, 失敗: %d, 合計: %d", successCount, failureCount, totalImages)
		// ---------------------------------
	}
}

// 添付画像をローカルに保存する関数
func downloadImage(url, filename string) error {
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
func uploadImageToDify(filename string) (string, error) {
	log.Printf("� Difyへのアップロード開始: %s", filename)

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
	mimeType := getMimeType(filename)

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
func runDifyWorkflowWithImage(fileID, userID, username string) (string, error) {
	log.Printf("� Difyワークフロー実行開始 - UserID: %s", userID)

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

	requestBody := map[string]interface{}{
		"inputs": map[string]interface{}{
			difyInputName: []interface{}{imageData},   // 配列形式で送信
			"payer":       fmt.Sprintf(`"%s"`, payer), // ユーザーに応じたpayer値
		},
		"response_mode": "blocking", // または "streaming"
		"user":          "discord-bot-user",
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("❌ JSONマーシャル失敗: %v", err)
		return "", fmt.Errorf("JSONマーシャルエラー: %v", err)
	}

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
		log.Printf("❌ ワークフロー実行失敗 - ステータス: %d", resp.StatusCode)

		// 400エラーの場合は入力パラメータの問題を指摘
		if resp.StatusCode == 400 {
			log.Printf("リクエストパラメータエラー: Difyワークフローの設定を確認してください")
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
