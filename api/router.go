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
)

var his []gogpt.ChatCompletionMessage

type Client struct {
	ctx      context.Context
	gpt      *gogpt.Client
	useModel string
	bouyomi  *bouyomichan.Client
	voicevox *voicevox.Client
}

func New(model string) *Client {
	v := voicevox.New()
	t, _ := v.GetSpeakers()
	for i := 0; i < len(t); i++ {
		fmt.Println(t[i])
	}

	return &Client{
		ctx:      context.Background(),
		gpt:      gogpt.NewClient(os.Getenv("CHATGPTTOKEN")),
		useModel: model,
		bouyomi:  bouyomichan.New("localhost:50001"),
		voicevox: v,
	}
}

func (p *Client) Router() {
	e := echo.New()
	e.Use(middleware.CORS())

	e.GET("/api/:message", p.RequestForStream)

	e.Logger.Fatal(e.Start(":8081"))
}
