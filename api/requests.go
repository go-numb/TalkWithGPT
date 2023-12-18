package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	gogpt "github.com/sashabaranov/go-openai"
)

const (
	MODELGPT4     = "gpt-4-1106-preview"
	MODELGPT4LONG = "gpt-4-1106-preview"

	MODELGPT3_5     = gogpt.GPT3Dot5Turbo
	MODELGPT3_5LONG = gogpt.GPT3Dot5Turbo16K
)

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

	// Token数に応じて、長文を扱うモデルに変える
	var temptext string
	for i := 0; i < len(his); i++ {
		temptext += his[i].Content
	}
	ids, _, _ := p.tokenizer.Encode(temptext)
	p._switchModel(len(ids))

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
		"code":   "success",
		"answer": fmt.Sprintf("%s\n\nmodel: %s, tokens: %d", res.Choices[0].Message.Content, p.UseModel(), len(ids)),
	})
}

func (p *Client) RequestForStream(c echo.Context) error {
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

	// Token数に応じて、長文を扱うモデルに変える
	var temptext string
	for i := 0; i < len(his); i++ {
		temptext += his[i].Content
	}
	ids, _, _ := p.tokenizer.Encode(temptext)
	p._switchModel(len(ids))

	req := gogpt.ChatCompletionRequest{
		Model:    p.UseModel(),
		Messages: his,
	}

	stream, err := p.gpt.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"code":   "error",
			"answer": err.Error(),
		})
	}
	defer stream.Close()

	// if stream.GetResponse().StatusCode != http.StatusOK {
	// 	return c.JSON(http.StatusBadRequest, map[string]any{
	// 		"code":   "error",
	// 		"answer": stream.GetResponse().Status,
	// 	})
	// }

	var (
		texts    []string
		readText string
	)
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"code":   "error",
				"answer": fmt.Sprintf("stream read error: %v", err),
			})
		}

		texts = append(texts, resp.Choices[0].Delta.Content)
		readText += resp.Choices[0].Delta.Content

		if _cutSmallPiece(readText) {
			go p.BouyomiSpeaking(readText)
			readText = ""
		}
	}

	if len([]rune(readText)) > 1 {
		go p.BouyomiSpeaking(readText)
	}
	readText = ""

	text := strings.Join(texts, "")
	his = append(his, gogpt.ChatCompletionMessage{
		Role:    "assistant",
		Content: text,
	})

	return c.JSON(http.StatusOK, map[string]any{
		"code":   "success",
		"answer": fmt.Sprintf("%s\n\nmodel: %s, tokens: %d", text, p.UseModel(), len(ids)),
	})
}

// 長い文章を句読点などで一時区切りする
var suffix = []string{"。", "！", "？"}

// _cutSmallPiece 長い文章を句読点などで一時区切りする
func _cutSmallPiece(q string) bool {
	for _, s := range suffix {
		if strings.HasSuffix(q, s) && len([]rune(q)) > 1 {
			return true
		}
	}
	return false
}

func (p *Client) _switchModel(n int) {
	if strings.HasPrefix(p.useModel, "gpt-4") {
		if n > 7500 {
			p.SetModel(MODELGPT4LONG)
		} else {
			p.SetModel(MODELGPT4)
		}
		return
	}

	// using base GPT3.5
	if n > 3500 {
		p.SetModel(MODELGPT3_5LONG)
	} else {
		p.SetModel(MODELGPT3_5)
	}

}
