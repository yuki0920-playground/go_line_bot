package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// ハンドラの呼び出し: ルーティングの役割
	http.HandleFunc("/", helloHandler)

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
