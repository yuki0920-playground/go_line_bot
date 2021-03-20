package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"unicode/utf8"

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

	replyMsg := getRestoInfo(lat, lng)
	res := linebot.NewTemplateMessage(
		"レストラン一覧",
		linebot.NewCarouselTemplate(replyMsg...).WithImageOptions("rectangle", "cover"),
	)

	if _, err := bot.ReplyMessage(e.ReplyToken, res).Do(); err != nil {
		log.Print(err)
	}
}

// レスポンスのJSONとのマッピングその1
type response struct {
	Results results `json: "results"`
}

// レスポンスのJSONとのマッピングその2
type results struct {
	Shop []shop `json:"shop"`
}

// レスポンスのJSONとのマッピングその3
type shop struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Photo   photo  `json:"photo"`
	URLS    urls   `json:"urls"`
}

// レスポンスのJSONとのマッピングその4
type photo struct {
	Mobile mobile `json:"mobile"`
}

// レスポンスのJSONとのマッピングその5
type mobile struct {
	L string `json:"l"`
}

// レスポンスのJSONとのマッピングその6
type urls struct {
	PC string `json:"pc"`
}

// 緯度と経度を受け取りカルーセルのカラムを配列で返す
func getRestoInfo(lat string, lng string) []*linebot.CarouselColumn {
	apikey := os.Getenv("HOTPEPPER_API_KEY")

	url := fmt.Sprintf(
		"https://webservice.recruit.co.jp/hotpepper/gourmet/v1/?format=json&key=%s&lat=%s&lng=%s",
		apikey,
		lat,
		lng,
	)
	log.Print("Request URL:", url)

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	// レスポンスが存在する場合は処理終了時に HTTP Response Body を閉じる
	defer resp.Body.Close()

	// レスポンスボディから読み込む
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var data response
	// jsonを構造体dataに変換する
	if err := json.Unmarshal(body, &data); err != nil {
		log.Fatal(err)
	}

	var ccs []*linebot.CarouselColumn
	for _, shop := range data.Results.Shop {
		addr := shop.Address
		// 住所は60字を超えるばあいは60字まで
		if 60 < utf8.RuneCountInString(addr) {
			addr = string([]rune(addr)[:60])
		}

		cc := linebot.NewCarouselColumn(
			shop.Photo.Mobile.L,
			shop.Name,
			addr,
			linebot.NewURIAction("ホットペッパーで開く", shop.URLS.PC),
		).WithImageOptions("#FFFFFF")
		ccs = append(ccs, cc)
	}

	return ccs
}
