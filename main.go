package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// HTTPサーバーを開始する関数
func startHTTPServer() *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/", HealthHandler)

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
	StartHealthCheckCron()

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
		// 現在のPayer判定結果も表示
		currentPayer := getPayerFromDiscordUser(m.Author.ID, m.Author.Username)
		userInfo := fmt.Sprintf("👤 **あなたの情報**\n```\nユーザーID: %s\nユーザー名: %s\n表示名: %s\n現在のPayer: %s\n```\n💡 この情報を使ってPayerを設定できます！",
			m.Author.ID, m.Author.Username, m.Author.GlobalName, currentPayer)
		_, _ = s.ChannelMessageSend(m.ChannelID, userInfo)

		// ログにも出力
		log.Printf("📋 !whoami実行 - UserID: %s, Username: %s, Payer: %s", m.Author.ID, m.Author.Username, currentPayer)
		return
	}

	// いくらコマンド
	if m.Content == "いくら" {
		// gasのurlを叩いて情報を取得し結果を返す。リクエストボディにパラメーターとしてaction:"get_latest_amount"を含める

		url := os.Getenv("GAS_ENDPOINT")
		data := `{"action":"get_latest_amount"}`

		bodyReader := strings.NewReader(data)

		resp, err := http.Post(url, "application/json", bodyReader)
		if err != nil {
			log.Printf("❌ POSTリクエストの送信中にエラーが発生しました: %v", err)
			s.ChannelMessageSend(m.ChannelID, "❌ データの取得に失敗しました")
			return
		}
		defer resp.Body.Close() // レスポンスボディを必ずクローズする

		// レスポンスボディを読み取る
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("❌ レスポンス読み取り失敗: %v", err)
			s.ChannelMessageSend(m.ChannelID, "❌ データの読み取りに失敗しました")
			return
		}

		// JSONをパース
		var result struct {
			Status       string   `json:"status"`
			Count        int      `json:"count"`
			CurrentMonth string   `json:"currentMonth"`
			Data         []string `json:"data"`
		}

		err = json.Unmarshal(respBody, &result)
		if err != nil {
			log.Printf("❌ JSONパース失敗: %v", err)
			s.ChannelMessageSend(m.ChannelID, "❌ データの解析に失敗しました")
			return
		}

		// Discordメッセージを作成（金額にカンマを追加）
		var message strings.Builder
		message.WriteString(fmt.Sprintf("**%sの記録**\n```\n", result.CurrentMonth))
		for _, item := range result.Data {
			// 金額にカンマを追加する処理
			formattedItem := FormatAmountWithComma(item)
			message.WriteString(formattedItem + "\n")
		}
		message.WriteString("```")

		// Discordに送信
		_, err = s.ChannelMessageSend(m.ChannelID, message.String())
		if err != nil {
			log.Printf("❌ メッセージ送信失敗: %v", err)
			return
		}

		log.Printf("� いくらコマンド実行成功 - UserID: %s", m.Author.ID)
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
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("📸 [%d/%d] %s を処理中...", i+1, len(m.Attachments), fileName))

			// 一時保存する場合（例: difyなどにPOST前にローカルで保持したい）
			err := DownloadImage(imageURL, fileName)
			if err != nil {
				log.Printf("❌ [%d/%d] 画像ダウンロード失敗 (%s): %v", i+1, len(m.Attachments), fileName, err)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ [%d/%d] %s のダウンロードに失敗しました: %v", i+1, len(m.Attachments), fileName, err))
				failureCount++
				continue
			}

			// 一時ディレクトリ内のファイルパスを取得
			tempFilePath := filepath.Join(os.TempDir(), fileName)

			// --- 画像を圧縮 ---
			compressedFileName, err := CompressImage(tempFilePath)
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
			fileID, err := UploadImageToDify(compressedFileName)
			if err != nil {
				log.Printf("❌ [%d/%d] Difyアップロード失敗 (%s): %v", i+1, len(m.Attachments), fileName, err)
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ [%d/%d] %s のDifyアップロードに失敗しました: %v", i+1, len(m.Attachments), fileName, err))
				os.Remove(compressedFileName)
				failureCount++
				continue
			}

			// 2. ワークフローを実行（画像を使用）
			result, err := RunDifyWorkflowWithImage(fileID, m.Author.ID, m.Author.Username)
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
					s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ [%d/%d] %s: Difyワークフローは実行されましたが、内部でエラーが発生しました。\n```\n%s\n```", i+1, len(m.Attachments), fileName, TruncateString(errorStr, 800)))
					failureCount++
				} else {
					// 正常な結果を表示
					// data.outputs.output配列から店舗・金額・項目を抽出
					var store, item string
					var amount int
					var display string

					// resultData["data"]["outputs"]["output"][0] を取得
					if data, ok := resultData["data"].(map[string]interface{}); ok {
						if outputs, ok := data["outputs"].(map[string]interface{}); ok {
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
											display = fmt.Sprintf("✅ [%d/%d] %s: Dify処理が完了しました！\n📍 店舗: %s\n💰 金額: %d円\n📝 項目: %s", i+1, len(m.Attachments), fileName, store, amount, item)
										}
									}
								}
							}
						}
					}

					if display != "" {
						s.ChannelMessageSend(m.ChannelID, display)
					} else {
						// パースできない場合は生のJSONを表示
						s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ [%d/%d] %s: Dify処理が完了しました！\n```json\n%s\n```", i+1, len(m.Attachments), fileName, TruncateString(result, 1200)))
					}
					successCount++
				}
			} else {
				// JSONパースできない場合はそのまま表示
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("✅ [%d/%d] %s: Dify処理が完了しました！\n```\n%s\n```", i+1, len(m.Attachments), fileName, TruncateString(result, 1200)))
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
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🎉 全ての画像処理が完了しました！\n✅ 成功: %d個\n", successCount))
		} else if successCount > 0 {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("⚠️ 一部の画像処理が完了しました。\n✅ 成功: %d個\n❌ 失敗: %d個", successCount, failureCount))
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("❌ 全ての画像処理が失敗しました。\n✅ 成功: %d個\n❌ 失敗: %d個", successCount, failureCount))
		}

		log.Printf("📊 画像処理サマリー - 成功: %d, 失敗: %d, 合計: %d", successCount, failureCount, totalImages)
		// ---------------------------------
	}
}
