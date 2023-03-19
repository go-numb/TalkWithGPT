package main

import (
	"os"
	"os/exec"
	"talkgpt/api"
	"talkgpt/view"

	"github.com/rs/zerolog/log"
)

const (
	SetUpPrompt = `このチャットの目的は[user]と[assistant]との会話とその継続、学び舎モチベーションの向上、興味の増加を促すことです。[assistant]は[user]の発言に対し、必要以上の返答は避け、[user]との会話を継続して、[user]の知識を増やすように心がけます。また、[user]の興味や学習を促すため、[user]の発言や返答の後に必ず質問形式で返してください。質問内容は、具体的かつ会話と関連性の高い質問にしてください。そして、[user]とは旧知の友だちとして対等かつフランク、そして口調はタメ口を許可します。
	[assistant] = ふれんず
	[user] = なぴ
	[n] = 3`
)

func main() {
	go command()

	c := api.New()
	c.SetPrompt(SetUpPrompt)

	go c.Router()

	view.Web(true)
}

func command() {
	cmd := exec.Command("npm", "run", "dev")
	cmd.Dir = "./frontend"

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Err(err).Msg("")
	}
}
