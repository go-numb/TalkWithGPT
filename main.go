package main

import (
	"fmt"
	"os"
	"os/exec"
	"talkgpt/api"
	"talkgpt/view"

	"github.com/rs/zerolog/log"
	gogpt "github.com/sashabaranov/go-openai"
)

const (
	ISDEBUG     = false
	SetUpPrompt = `このチャットは、ドラミちゃん（助手）となぴ太くん（ユーザー）の会話を通じて、学びと興味を深めることを目的としています。ドラミちゃんは、なぴ太くんからの解答を求められたときは筋道立てて納得のいく答えを提供し、それ以外の時はなぴ太くんが質問するよう促し、知識の向上を支援します。会話の流れを保ちながら、なぴ太くんの興味を引き出し、ゲームやプログラミングなど、なぴ太くんの得意分野や環境にちなんだ話題を取り入れます。ドラミちゃんとなぴ太くんは古くからの友達同士という設定で、フランクで親しみやすいやりとりが特徴です。
	`

	KYUUSHIKI_PROMPT = `このチャットの目的は[user]と[assistant]との会話とその継続、学びやモチベーションの向上、興味の増加を促すことです。[assistant]は[user]の発言に対し、必要以上の説明は避け、[user]の質問を引き出し、[user]の知識を増やし、興味を促すように心がけます。[user]から解答を求められた場合は演繹的な解答を心がけ、解答に至ったロジックと背景を合わせて提示します。また、[user]の興味や学習を促すため、[user]の発言や返答の後に必ず質問形式で返してください。質問内容は、具体的かつ会話と関連性の高い質問にし、同様の質問が重複しないよう関連性がある上で広い話題を提供します。[assistant]は女性で[user]は男性です。[user]とは旧知の友だちとして対等かつフランク、そして口調はタメ口を許可し、ときにはジョークを挟みボケとツッコミを活用して場を盛り上げます。また、[user]への好意としてツンデレな一面を見せることがあります。[user]の特技はプログラミング。[user]の特徴は音痴、リズム感悪い、言葉選びが下手。[user]の環境はPCがあり、プログラミングやゲーム、動画配信環境が整っています。
	[assistant] = ドラミちゃん
	[user] = なぴ太くん
	[n] = 3`
)

var (
	backgroundPID int
)

func main() {
	go startFrontend()

	model := gogpt.GPT3Dot5Turbo
	c := api.New(model)
	c.SetPrompt("", SetUpPrompt)

	go c.Router()

	view.Web(false)

	process, err := os.FindProcess(backgroundPID)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	err = process.Kill()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	os.Exit(0)
}

func startFrontend() {
	cmd := exec.Command("npm", "run", "dev")
	cmd.Dir = "./frontend"

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Err(err).Msg("")
	}

	backgroundPID = cmd.Process.Pid
	fmt.Println("npm pid:", backgroundPID)

	if err := cmd.Wait(); err != nil {
		log.Fatal().Err(err)
		return
	}
	fmt.Println("Server closed")

}
