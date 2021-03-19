package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	// LINE SDK
	"github.com/line/line-bot-sdk-go/linebot"
	// 環境変数
	"github.com/joho/godotenv"
)

func main() {
	// ハンドラの呼び出し: ルーティングの役割
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/callback", lineHandler)

	// 標準出力
	fmt.Println("http://localhost:8080 で起動")

	// ログ出力とHTTPサーバーの起動
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// w は http.ResponseWriter型の引数 r は http.Request型のポインタの引数
func helloHandler(w http.ResponseWriter, r *http.Request) {
	msg := "Hello World"
	fmt.Fprint(w, msg)
}

func lineHandler(w http.ResponseWriter, r *http.Request) {
	log.Print("リクエスト受信")

	envLoad()

	secretKey := os.Getenv("LINE_SECRET_KEY")
	secretToken := os.Getenv("LINE_ACCES_TOKEN")
	fmt.Println("LINE_SECRET_KEY", secretKey)
	fmt.Println("LINE_ACCES_TOKEN", secretToken)

	// botを初期化
	bot, err := linebot.New(
		secretKey,
		secretToken,
	)

	// エラーに値があればログに出力する
	if err != nil {
		log.Fatal(err)
	}
	// botのParseRequestメソッドの引数にlineHandlerを渡す
	events, err := bot.ParseRequest(r)
	if err != nil {
		// エラーがある場合エラーの種類によってレスポンスコードを出し分ける
		if err == linebot.ErrInvalidSignature {
			// Bad Request
			w.WriteHeader(400)
		} else {
			// Internal Server Error
			w.WriteHeader(500)
		}
	}
	// インデックスは不要なので_に入れてループする
	for _, event := range events {
		// メッセージタイプのイベントならば
		if event.Type == linebot.EventTypeMessage {
			// メッセージを変数に入れてメッセージの種類ごとに判定する
			switch message := event.Message.(type) {
			case *linebot.TextMessage: // テキストメッセージの場合
				replyMessage := message.Text
				// 返信用トークンをつけてメッセージを送信する
				_, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do()
				if err != nil {
					log.Print(err)
				}
			case *linebot.LocationMessage: // 位置情報の場合
				sendRestoInfo(bot, event)
			}
		}
	}
}

func envLoad() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func sendRestoInfo(bot *linebot.Client, e *linebot.Event) {
	// e.Message の型が *linebot.LocationMessage であるかのチェック(型アサーション)をしつつ変数に代入
	msg := e.Message.(*linebot.LocationMessage)

	lat := strconv.FormatFloat(msg.Latitude, 'f', 2, 64)
	lng := strconv.FormatFloat(msg.Longitude, 'f', 2, 64)

	replyMsg := fmt.Sprintf("緯度: %s\n軽度: %s", lat, lng)

	_, err := bot.ReplyMessage(e.ReplyToken, linebot.NewTextMessage(replyMsg)).Do()
	if err != nil {
		log.Print(err)
	}
}
