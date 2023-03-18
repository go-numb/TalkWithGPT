# TALKWITHGPT
CHATGPTとお話しするアプリです。
ブラウザ、もしくはDesktopAppからマイク音声を入力するとテキスト化し、ChatGPTAPIへ問い合わせ。返り値を音声出力します。

現在、クリックしてから音声入力スタートです。  
今後、常々入力しておき、音声コマンドから問い合わせテキストを作るような仕組みにしたい。

ゲーム中に質問したり、会話に付き合ってもらったり、雑学教えてもらったりしたい。そういう要望の末書いたアプリ。そう、これは、ほぼドラえもん。

VoiceVoxを導入すれば、可愛い声でも喋ります。

## Usage
```
$ git clone this repo
$ cd thisDir/frontend && npm i
$ npm run dev
$ cd ../
$ export CHATGPTTOKEN="xxxxxxxxxxxxxx"
$ go run main.go

```

# Author
[@_numbP](https://twitter.com/_numbP)