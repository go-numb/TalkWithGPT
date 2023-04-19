package api

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	myprompts "github.com/go-numb/my-prompts"
	"github.com/rs/zerolog/log"
	gogpt "github.com/sashabaranov/go-openai"
)

func (p *Client) SetPrompt(key, prompt string) {
	if key != "" {
		prompts := myprompts.Map()
		fmt.Println("Prompts list")
		var i int
		for k := range prompts {
			fmt.Printf("%d: %s\n", i, k)
			i++
		}
		prompt = prompts[key].Command
	}

	ctx, cancel := context.WithTimeout(p.ctx, 60*time.Second)
	defer cancel()

	his = append(his, gogpt.ChatCompletionMessage{
		Role:    "system",
		Content: prompt,
	})
	req := gogpt.ChatCompletionRequest{
		Model:    p.UseModel(),
		Messages: his,
	}

	res, err := p.gpt.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Err(err).Msg("request error")
		return
	}

	if len(res.Choices) == 0 {
		log.Error().Msg("has not choices")
		return
	}

	his = append(his, gogpt.ChatCompletionMessage{
		Role:    "assistant",
		Content: res.Choices[0].Message.Content,
	})
	log.Info().Msgf("set prompt: %s, assistant says %s", prompt, res.Choices[0].Message.Content)
}

func (p *Client) BouyomiSpeaking(s string) {
	p.bouyomi.Speed = 88
	p.bouyomi.Tone = 105
	// p.bouyomi.Voice = bouyomichan.VoiceDefault
	p.bouyomi.Voice = 10027 // Voicevox 冥鳴ひまり ノーマル
	p.bouyomi.Volume = 110
	if err := p.bouyomi.Speaking(s); err != nil {
		log.Err(err).Msg("")
		return
	}

	ctx, cacel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cacel()

L:
	for {
		select {
		case <-ctx.Done():
			if err := p.bouyomi.Stop(); err != nil {
				log.Err(fmt.Errorf("[ERROR] bouyomi stoped error, %v", err)).Msg("")
			}
			break L
		default:
			if !p.bouyomi.IsNowPlayng() {
				break L
			}
			time.Sleep(time.Second)
		}
	}
}

// VoxSpeaking メモリ使用しまくり
func (p *Client) VoxSpeaking(s string) {
	params, err := p.voicevox.GetQuery(0, s)
	if err != nil {
		log.Err(err).Msg("")
		return
	}

	params.VolumeScale = 0.3

	p.voicevox.Set(params)
	// fmt.Printf("%#v\n", params)

	b, err := p.voicevox.Synth(14, params)
	if err != nil {
		log.Err(err).Msg("")
		return
	}

	if err := p.voicevox.Speaking(params, b[44:]); err != nil {
		log.Err(err).Msg("")
		return
	}
}

func (p *Client) SetModel(q string) {
	p.useModel = q
}

func (p *Client) UseModel() string {
	return p.useModel
}

func Save(histories []gogpt.ChatCompletionMessage) (string, error) {
	if histories == nil || len(his) < 1 {
		return "保存する会話履歴がありません", fmt.Errorf("[ERROR] has not histories")
	}

	var (
		title string
		l     int = len(histories) - 1

		contents = make([]string, len(histories))
	)
	if len([]rune(histories[l].Content)) > 5 {
		title = string([]rune(his[l].Content)[:5])
	}
	f, _ := os.Create(fmt.Sprintf("./_data/memo-%s-%s.md", time.Now().Format("200601021504"), title))
	defer f.Close()

	for i := 0; i < len(histories); i++ {
		if histories[i].Role == "system" {
			continue
		}
		contents[i] = histories[i].Content
	}

	f.WriteString(strings.Join(contents, "\n\n"))

	return "保存しました", nil
}
