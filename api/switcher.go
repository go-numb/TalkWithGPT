package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	gogpt "github.com/sashabaranov/go-openai"
)

func (p *Client) Switcher(c echo.Context, message string) (isSwitch bool, err error) {
	isSwitch = true
	q := strings.ToLower(message)

	// Has not query
	if message == "" {
		return isSwitch, c.JSON(http.StatusBadRequest, map[string]any{
			"code":  "error bad request, has not request message",
			"error": errors.New("has not query"),
		})
	}

	// Reset...
	if strings.HasPrefix(q, "reset") ||
		strings.HasPrefix(q, "リセット") ||
		strings.HasPrefix(q, "りせっと") ||
		strings.HasPrefix(q, "履歴リセット") ||
		strings.HasPrefix(q, "履歴りせっと") {
		// System sayを残す
		temp := make([]gogpt.ChatCompletionMessage, 0)
		for i := 0; i < len(his); i++ {
			if his[i].Role != "system" {
				continue
			}
			temp = append(temp, his[i])
		}

		// System command promptsを残す
		his = temp
		return isSwitch, c.JSON(http.StatusOK, map[string]any{
			"code":  "success, reset histories",
			"error": "success",
		})
	}

	// Change model
	if strings.HasPrefix(q, "モデル") {
		temp := strings.Replace(q, "モデル", "", 1)
		if strings.HasPrefix(temp, "3.5") {
			p.SetModel(gogpt.GPT3Dot5Turbo0301)
			go p.BouyomiSpeaking("モデル3.5に切り替え。")
			return isSwitch, c.JSON(http.StatusOK, map[string]any{
				"code":   "success",
				"answer": "success, set model: " + gogpt.GPT3Dot5Turbo0301,
			})
		} else if strings.HasPrefix(temp, "4.0") {
			p.SetModel(gogpt.GPT40314)
			go p.BouyomiSpeaking("モデル4.0に切り替え。")
			return isSwitch, c.JSON(http.StatusOK, map[string]any{
				"code":   "success",
				"answer": "success, set model: " + gogpt.GPT3Dot5Turbo0301,
			})
		}
	}

	// Memo...
	if strings.HasPrefix(q, "メモ") {
		temp := strings.Replace(q, "メモ", "", 1)
		// 履歴の読み込み
		if strings.HasPrefix(temp, "読み込み") {
			temp = strings.Replace(temp, "読み込み", "", 1)
			temp = strings.Replace(temp, "。", "", -1)
			temp = strings.Replace(temp, "、", "", -1)
			text, err := _search(temp)
			if err != nil {
				return isSwitch, c.JSON(http.StatusOK, map[string]any{
					"code":   "error",
					"answer": err.Error(),
				})
			}
			go p.BouyomiSpeaking(fmt.Sprintf("「%s」を含むファイルを見つけたよ。", temp))

			return isSwitch, c.JSON(http.StatusOK, map[string]any{
				"code":   "success",
				"answer": text,
			})
		}

		// 履歴の書き込み
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
	}

	return false, nil
}

var Dir = "./_data"

// _search ファイル走査 該当のサーチクエリが含まれる先頭のファイルを返す単純な関数
func _search(q string) (string, error) {
	files, _ := os.ReadDir(Dir)
	for _, file := range files {
		f, _ := os.Open(filepath.Join(Dir, file.Name()))
		buf, err := io.ReadAll(f)
		if err != nil {
			log.Err(err).Msg("")
			continue
		}
		f.Close()

		if strings.Contains(string(buf), q) {
			return string(buf), nil
		}
	}

	return "has not file", nil
}
