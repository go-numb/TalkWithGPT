package api

import (
	"context"
	"fmt"
	"os"

	"github.com/go-numb/go-bouyomichan"
	"github.com/go-numb/go-voicevox"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	gogpt "github.com/sashabaranov/go-openai"
	"github.com/tiktoken-go/tokenizer"

	"github.com/rs/zerolog/log"
)

var his []gogpt.ChatCompletionMessage

type Client struct {
	ctx       context.Context
	gpt       *gogpt.Client
	tokenizer tokenizer.Codec
	useModel  string
	bouyomi   *bouyomichan.Client
	voicevox  *voicevox.Client
}

func New(model string) *Client {
	v := voicevox.New()
	t, _ := v.GetSpeakers()
	for i := 0; i < len(t); i++ {
		fmt.Println(t[i])
	}

	enc, err := tokenizer.Get(tokenizer.P50kEdit)
	if err != nil {
		log.Fatal().Msgf("gpt tokenizer, %s", err)
	}

	return &Client{
		ctx:       context.Background(),
		gpt:       gogpt.NewClient(os.Getenv("CHATGPTTOKEN")),
		tokenizer: enc,
		useModel:  model,
		bouyomi:   bouyomichan.New("localhost:50001"),
		voicevox:  v,
	}
}

func (p *Client) Router() {
	e := echo.New()
	e.Use(middleware.CORS())

	e.GET("/api/:message", p.RequestForStream)

	e.Logger.Fatal(e.Start("localhost:8081"))
}
