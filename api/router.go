package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-numb/go-bouyomichan"
	"github.com/go-numb/go-voicevox"
	myprompts "github.com/go-numb/my-prompts"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	gogpt "github.com/sashabaranov/go-openai"
)

type Client struct {
	ctx      context.Context
	gpt      *gogpt.Client
	useModel string
	bouyomi  *bouyomichan.Client
	voicevox *voicevox.Client
}

func New() *Client {
	v := voicevox.New()
	t, _ := v.GetSpeakers()
	for i := 0; i < len(t); i++ {
		fmt.Println(t[i])
	}

	return &Client{
		ctx:      context.Background(),
		gpt:      gogpt.NewClient(os.Getenv("CHATGPTTOKEN")),
		useModel: gogpt.GPT3Dot5Turbo0301,
		bouyomi:  bouyomichan.New("localhost:50001"),
		voicevox: v,
	}
}

func (p *Client) Router() {
	e := echo.New()
	e.Use(middleware.CORS())

	e.GET("/api/:message", p.Request)

	e.Logger.Fatal(e.Start(":8081"))
}

var his []gogpt.ChatCompletionMessage

func (p *Client) Request(c echo.Context) error {
	message := c.Param("message")
	fmt.Println(message)

	if isSwitch, err := p.Switcher(c, message); isSwitch {
		return err
	}

	ctx, cancel := context.WithTimeout(p.ctx, 180*time.Second)
	defer cancel()

	his = append(his, gogpt.ChatCompletionMessage{
		Role:    "user",
		Content: message,
	})
	req := gogpt.ChatCompletionRequest{
		Model:    p.UseModel(),
		Messages: his,
	}

	res, err := p.gpt.CreateChatCompletion(ctx, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"code":   "error",
			"answer": err.Error(),
		})
	}

	if len(res.Choices) <= 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"code":   "error",
			"answer": "nothing answer",
		})
	}

	his = append(his, gogpt.ChatCompletionMessage{
		Role:    "assistant",
		Content: res.Choices[0].Message.Content,
	})

	go p.BouyomiSpeaking(res.Choices[0].Message.Content)
	// go p.VoxSpeaking(res.Choices[0].Message.Content)

	return c.JSON(http.StatusOK, map[string]any{
		"answer": res.Choices[0].Message.Content,
	})
}

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
	p.bouyomi.Volume = 70
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

func (p *Client) Switcher(c echo.Context, message string) (isSwitch bool, err error) {
	isSwitch = true
	q := strings.ToLower(message)
	if q == "reset" {
		// System command promptsを残す
		his = []gogpt.ChatCompletionMessage{
			his[0],
		}
		return isSwitch, c.JSON(http.StatusOK, map[string]any{
			"code":  "success, reset histories",
			"error": "success",
		})
	} else if q == "リセット" {
		his = []gogpt.ChatCompletionMessage{
			his[0],
		}
		return isSwitch, c.JSON(http.StatusOK, map[string]any{
			"code":   "success",
			"answer": "success, reset histories",
		})
	} else if strings.HasPrefix(q, "履歴リセット") {
		his = []gogpt.ChatCompletionMessage{
			his[0],
		}
		return isSwitch, c.JSON(http.StatusOK, map[string]any{
			"code":   "success",
			"answer": "success, reset histories",
		})
	} else if strings.HasPrefix(q, "モデル3.5") {
		p.SetModel(gogpt.GPT3Dot5Turbo0301)
		return isSwitch, c.JSON(http.StatusOK, map[string]any{
			"code":   "success",
			"answer": "success, set model: " + gogpt.GPT3Dot5Turbo0301,
		})
	} else if strings.HasPrefix(q, "モデル4.0") {
		p.SetModel(gogpt.GPT40314)
		return isSwitch, c.JSON(http.StatusOK, map[string]any{
			"code":   "success",
			"answer": "success, set model: " + gogpt.GPT40314,
		})
	} else if strings.HasPrefix(q, "メモ") {
		say, err := Save(his)
		if err != nil {
			return isSwitch, c.JSON(http.StatusOK, map[string]any{
				"code":   "error",
				"answer": err.Error(),
			})
		}
		go p.BouyomiSpeaking(say)

		return isSwitch, c.JSON(http.StatusOK, map[string]any{
			"code":   "success",
			"answer": say,
		})
	} else if message == "" {
		return isSwitch, c.JSON(http.StatusBadRequest, map[string]any{
			"code":  "error bad request, has not request message",
			"error": errors.New("has not query"),
		})
	}

	return false, nil
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
	f, _ := os.Create(fmt.Sprintf("./_data/memo-%s-%s.txt", time.Now().Format("200601021504"), title))
	defer f.Close()

	for i := 0; i < len(histories); i++ {
		if histories[i].Role == "system" {
			continue
		}
		contents[i] = histories[i].Content
	}

	f.WriteString(strings.Join(contents, "\n"))

	return "保存しました", nil
}
