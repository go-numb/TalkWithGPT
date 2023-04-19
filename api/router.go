package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

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

func New() *Client {
	v := voicevox.New()
	t, _ := v.GetSpeakers()
	for i := 0; i < len(t); i++ {
		fmt.Println(t[i])
	}

	return &Client{
		ctx:      context.Background(),
		gpt:      gogpt.NewClient(os.Getenv("CHATGPTTOKEN")),
		useModel: gogpt.GPT40314,
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
