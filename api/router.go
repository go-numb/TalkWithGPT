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
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	gogpt "github.com/sashabaranov/go-openai"
)

type Client struct {
	ctx      context.Context
	gpt      *gogpt.Client
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
		bouyomi:  bouyomichan.New("localhost:50001"),
		voicevox: v,
	}
}

func (p *Client) Router() {
	e := echo.New()
	e.Use(middleware.CORS())

	e.GET("/api/:message", p.Request)

	e.Logger.Fatal(e.Start(":8080"))
}

var his []gogpt.ChatCompletionMessage

func (p *Client) Request(c echo.Context) error {
	message := c.Param("message")
	fmt.Println(message)

	q := strings.ToLower(message)
	if q == "reset" {
		his = []gogpt.ChatCompletionMessage{}
		return c.JSON(http.StatusOK, map[string]any{
			"code":  "success, reset histories",
			"error": "success",
		})
	} else if q == "リセット" {
		his = []gogpt.ChatCompletionMessage{}
		return c.JSON(http.StatusOK, map[string]any{
			"code":  "success, reset histories",
			"error": "success",
		})
	} else if strings.HasPrefix(q, "履歴リセット") {
		his = []gogpt.ChatCompletionMessage{}
		return c.JSON(http.StatusOK, map[string]any{
			"code":  "success, reset histories",
			"error": "success",
		})
	} else if message == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"code":  "error bad request, has not request message",
			"error": errors.New("has not query"),
		})
	}

	ctx, cancel := context.WithTimeout(p.ctx, 60*time.Second)
	defer cancel()

	his = append(his, gogpt.ChatCompletionMessage{
		Role:    "user",
		Content: message,
	})
	req := gogpt.ChatCompletionRequest{
		Model:    gogpt.GPT3Dot5Turbo,
		Messages: his,
	}

	res, err := p.gpt.CreateChatCompletion(ctx, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"code":  "error",
			"error": err.Error(),
		})
	}

	his = append(his, gogpt.ChatCompletionMessage{
		Role:    "assistant",
		Content: res.Choices[0].Message.Content,
	})

	go p.BouyomiSpeaking(res.Choices[0].Message.Content)
	go p.VoxSpeaking(res.Choices[0].Message.Content)

	return c.JSON(http.StatusOK, map[string]any{
		"answer": res.Choices[0].Message.Content,
	})
}

func (p *Client) SetPrompt(q string) {
	ctx, cancel := context.WithTimeout(p.ctx, 60*time.Second)
	defer cancel()

	his = append(his, gogpt.ChatCompletionMessage{
		Role:    "system",
		Content: q,
	})
	req := gogpt.ChatCompletionRequest{
		Model:    gogpt.GPT3Dot5Turbo,
		Messages: his,
	}

	res, err := p.gpt.CreateChatCompletion(ctx, req)
	if err != nil {
		log.Err(err).Msg("request error")
	}

	if len(res.Choices) == 0 {
		log.Error().Msg("has not choices")
		return
	}

	his = append(his, gogpt.ChatCompletionMessage{
		Role:    "assistant",
		Content: res.Choices[0].Message.Content,
	})
	log.Info().Msg(res.Choices[0].Message.Content)
}

func (p *Client) BouyomiSpeaking(s string) {
	p.bouyomi.Speed = 110
	p.bouyomi.Tone = 120
	p.bouyomi.Voice = bouyomichan.VoiceDefault
	p.bouyomi.Volume = 60
	if err := p.bouyomi.Speaking(s); err != nil {
		log.Err(err).Msg("")
		return
	}

	ctx, cacel := context.WithTimeout(context.Background(), 60*time.Second)
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
				break
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

	params.VolumeScale = 0.4

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
