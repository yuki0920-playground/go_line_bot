# GO でつくる LINE BOT

## セットアップ
https://developers.line.biz/console

- `ngrok`起動
```
ngrok http 8080
```

- ForwardingのhttpsのURLをコピー
  ex: `https://5758effb6ed1.ngrok.io`

- URLをLINE[デベロッパーコンソール](https://developers.line.biz/console/)のMessaging APIのWebhook URLに`/callback`を追加し設定する
  ex: `https://5758effb6ed1.ngrok.io/callback`

- main.goを起動

```
go run main.go
```

## エンドポイント
- `/` `Hello, World` を表示
- `/callback` LINEBOT 用
